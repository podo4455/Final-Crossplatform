package analysis

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// VisionAnalyzer เรียก Google Cloud Vision API แบบ REST ด้วย net/http
// สามารถ override Endpoint ได้เพื่อใช้กับ httptest ในการเทส
type VisionAnalyzer struct {
	APIKey   string
	Client   *http.Client
	Endpoint string
	Timeout  time.Duration
}

// โครงสร้าง request สำหรับ Google Vision API
type visionRequest struct {
	Requests []visionRequestItem `json:"requests"`
}

type visionRequestItem struct {
	Image    visionImage     `json:"image"`
	Features []visionFeature `json:"features"`
}

type visionImage struct {
	Content string `json:"content"`
}

type visionFeature struct {
	Type       string `json:"type"`
	MaxResults int    `json:"maxResults"`
}

// โครงสร้าง response สำหรับ Google Vision API
type visionResponse struct {
	Responses []visionResponseItem `json:"responses"`
}

type visionResponseItem struct {
	LabelAnnotations []visionLabelAnnotation `json:"labelAnnotations"`
}

type visionLabelAnnotation struct {
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	Topicality  float64 `json:"topicality"`
}

// DetectLabels อ่านไฟล์รูป แปลงเป็น base64 แล้วส่งไปยัง Google Cloud Vision API เพื่อขอ Label Detection
// ไม่กรอง score ที่นี่ ให้ service เป็นคนตัดสินใจตาม threshold
func (v VisionAnalyzer) DetectLabels(ctx context.Context, imagePath string) ([]Label, error) {
	if imagePath == "" {
		return nil, fmt.Errorf("ต้องระบุตำแหน่งไฟล์รูป")
	}

	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("หมดเวลาการเชื่อมต่อ กรุณาลองใหม่อีกครั้ง")
		}
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, err
	}

	// อ่านไฟล์รูป
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("อ่านไฟล์รูปไม่สำเร็จ: %w", err)
	}

	// แปลงเป็น base64
	encoded := base64.StdEncoding.EncodeToString(data)

	// สร้าง request body
	reqBody := visionRequest{
		Requests: []visionRequestItem{
			{
				Image: visionImage{
					Content: encoded,
				},
				Features: []visionFeature{
					{
						Type:       "LABEL_DETECTION",
						MaxResults: 10,
					},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("สร้าง request body ไม่สำเร็จ: %w", err)
	}

	// กำหนด endpoint
	endpoint := v.Endpoint
	if endpoint == "" {
		if v.APIKey == "" {
			return nil, fmt.Errorf("ยังไม่ได้ตั้ง API key สำหรับ Google Cloud Vision")
		}
		endpoint = fmt.Sprintf("https://vision.googleapis.com/v1/images:annotate?key=%s", v.APIKey)
	}

	// กำหนด http client
	client := v.Client
	if client == nil {
		timeout := v.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}

	// ส่ง request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("สร้าง request ไม่สำเร็จ: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("หมดเวลาการเชื่อมต่อ Google Cloud Vision กรุณาลองใหม่อีกครั้ง")
		}

		var netErr *net.OpError
		if errors.As(err, &netErr) {
			return nil, fmt.Errorf("ไม่มีการเชื่อมต่ออินเทอร์เน็ต ไม่สามารถวิเคราะห์รูปด้วย Google Cloud Vision ได้")
		}

		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "no such host") ||
			strings.Contains(errMsg, "network is unreachable") ||
			strings.Contains(errMsg, "i/o timeout") ||
			strings.Contains(errMsg, "temporary failure in name resolution") ||
			strings.Contains(errMsg, "dial tcp") {
			return nil, fmt.Errorf("ไม่มีการเชื่อมต่ออินเทอร์เน็ต ไม่สามารถวิเคราะห์รูปด้วย Google Cloud Vision ได้")
		}

		return nil, fmt.Errorf("เรียก Google Vision API ไม่สำเร็จ: %w", err)
	}
	defer resp.Body.Close()

	// ตรวจสอบ status code
	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("Google Vision API ตอบกลับด้วยสถานะ %d: %s", resp.StatusCode, string(errBody))
	}

	// Parse response
	var visionResp visionResponse
	if err := json.NewDecoder(resp.Body).Decode(&visionResp); err != nil {
		return nil, fmt.Errorf("แปลง response จาก Google Vision ไม่สำเร็จ: %w", err)
	}

	if len(visionResp.Responses) == 0 {
		return []Label{}, nil
	}

	// แปลงเป็น []Label
	labels := make([]Label, 0, len(visionResp.Responses[0].LabelAnnotations))
	for _, ann := range visionResp.Responses[0].LabelAnnotations {
		labels = append(labels, Label{
			Description: ann.Description,
			Score:       ann.Score,
		})
	}

	return labels, nil
}

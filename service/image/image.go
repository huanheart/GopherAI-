package image

import (
	"GopherAI/common/image"
	"io"
	"mime/multipart"
)

// RecognizeImage a new recognizer for each request, handles the image recognition.
func RecognizeImage(file *multipart.FileHeader) (string, error) {
	// Hardcoded paths and parameters as requested
	modelPath := "/root/models/mobilenetv2/mobilenetv2-7.onnx"
	labelPath := "/root/imagenet_classes.txt"
	inputH, inputW := 224, 224

	// Create a new recognizer for this specific request, using the updated function signature
	recognizer, err := image.NewImageRecognizer(modelPath, labelPath, inputH, inputW)
	if err != nil {
		return "", err
	}
	defer recognizer.Close() // Ensure resources are released

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	// Perform prediction
	return recognizer.PredictFromBuffer(buf)
}

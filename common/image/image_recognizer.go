package image

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"

	ort "github.com/yalue/onnxruntime_go"
	"gocv.io/x/gocv"
)

type ImageRecognizer struct {
	session      *ort.Session[float32]
	inputName    string
	outputName   string
	inputH       int
	inputW       int
	labels       []string
	inputTensor  *ort.Tensor[float32]
	outputTensor *ort.Tensor[float32]
}

const (
	defaultInputName  = "input"
	defaultOutputName = "output"
)

// NewImageRecognizer 创建识别器（自动使用默认 input/output 名称）
func NewImageRecognizer(modelPath, labelPath string, inputH, inputW int) (*ImageRecognizer, error) {
	if inputH <= 0 || inputW <= 0 {
		inputH, inputW = 224, 224
	}

	// 初始化 ONNX 环境（全局一次）
	if err := ort.InitializeEnvironment(); err != nil {
		return nil, fmt.Errorf("onnxruntime initialize error: %w", err)
	}

	// 预先创建输入输出 Tensor
	inputShape := ort.NewShape(1, 3, int64(inputH), int64(inputW))
	inData := make([]float32, inputShape.FlattenedSize())
	inTensor, err := ort.NewTensor(inputShape, inData)
	if err != nil {
		return nil, fmt.Errorf("create input tensor failed: %w", err)
	}

	outShape := ort.NewShape(1, 1000)
	outTensor, err := ort.NewEmptyTensor[float32](outShape)
	if err != nil {
		inTensor.Destroy()
		return nil, fmt.Errorf("create output tensor failed: %w", err)
	}

	// 创建 Session
	session, err := ort.NewSession[float32](
		modelPath,
		[]string{defaultInputName},
		[]string{defaultOutputName},
		[]*ort.Tensor[float32]{inTensor},
		[]*ort.Tensor[float32]{outTensor},
	)
	if err != nil {
		inTensor.Destroy()
		outTensor.Destroy()
		return nil, fmt.Errorf("create onnx session failed: %w", err)
	}

	// 读取 label 文件
	labels, err := loadLabels(labelPath)
	if err != nil {
		session.Destroy()
		inTensor.Destroy()
		outTensor.Destroy()
		return nil, err
	}

	return &ImageRecognizer{
		session:      session,
		inputName:    defaultInputName,
		outputName:   defaultOutputName,
		inputH:       inputH,
		inputW:       inputW,
		labels:       labels,
		inputTensor:  inTensor,
		outputTensor: outTensor,
	}, nil
}

func (r *ImageRecognizer) Close() {
	if r.session != nil {
		_ = r.session.Destroy()
		r.session = nil
	}
	if r.inputTensor != nil {
		_ = r.inputTensor.Destroy()
		r.inputTensor = nil
	}
	if r.outputTensor != nil {
		_ = r.outputTensor.Destroy()
		r.outputTensor = nil
	}
}

func (r *ImageRecognizer) PredictFromFile(imagePath string) (string, error) {
	if _, err := os.Stat(imagePath); err != nil {
		return "", fmt.Errorf("image not found: %w", err)
	}
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return "", fmt.Errorf("failed to read image: %s", imagePath)
	}
	defer img.Close()
	return r.PredictFromMat(img)
}

func (r *ImageRecognizer) PredictFromBuffer(buf []byte) (string, error) {
	mat, err := gocv.IMDecode(buf, gocv.IMReadColor)
	if err != nil {
		return "", fmt.Errorf("imdecode failed: %w", err)
	}
	defer mat.Close()
	if mat.Empty() {
		return "", errors.New("decoded image is empty")
	}
	return r.PredictFromMat(mat)
}

func (r *ImageRecognizer) PredictFromMat(img gocv.Mat) (string, error) {
	if img.Empty() {
		return "", errors.New("input image is empty")
	}

	dst := gocv.NewMat()
	defer dst.Close()
	gocv.Resize(img, &dst, image.Pt(r.inputW, r.inputH), 0, 0, gocv.InterpolationDefault)

	h, w, ch := r.inputH, r.inputW, 3
	data := make([]float32, h*w*ch)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := dst.GetVecbAt(y, x)
			rv := float32(v[2]) / 255.0
			gv := float32(v[1]) / 255.0
			bv := float32(v[0]) / 255.0

			data[y*w+x] = rv
			data[h*w+y*w+x] = gv
			data[2*h*w+y*w+x] = bv
		}
	}

	inData := r.inputTensor.GetData()
	copy(inData, data)

	if err := r.session.Run(); err != nil {
		return "", fmt.Errorf("onnx run error: %w", err)
	}

	outData := r.outputTensor.GetData()
	if len(outData) == 0 {
		return "", errors.New("empty output from model")
	}

	maxIdx := 0
	maxVal := outData[0]
	for i := 1; i < len(outData); i++ {
		if outData[i] > maxVal {
			maxVal = outData[i]
			maxIdx = i
		}
	}

	if maxIdx >= 0 && maxIdx < len(r.labels) {
		return r.labels[maxIdx], nil
	}
	return "Unknown", nil
}

func loadLabels(path string) ([]string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("open label file failed: %w", err)
	}
	defer f.Close()

	var labels []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line != "" {
			labels = append(labels, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read labels failed: %w", err)
	}
	if len(labels) == 0 {
		return nil, fmt.Errorf("no labels found in %s", path)
	}
	return labels, nil
}

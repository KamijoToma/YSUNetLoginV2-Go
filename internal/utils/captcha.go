package utils

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	_ "image/gif"
	_ "image/png"
)

// CaptchaDisplayMode defines how captcha should be displayed
type CaptchaDisplayMode string

const (
	DisplayASCII CaptchaDisplayMode = "ascii"
	DisplayFile  CaptchaDisplayMode = "file"
	DisplayBoth  CaptchaDisplayMode = "both"
)

// ImageToASCII converts image data to ASCII art
func ImageToASCII(imageData io.Reader, width int, charSet string) (string, error) {
	// Character sets for different densities
	var asciiChars string
	switch charSet {
	case "dense":
		asciiChars = "@%#*+=-:. "
	case "extended":
		asciiChars = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,\"^`'. "
	default: // standard
		asciiChars = "@#S%?*+;:,. "
	}

	// Decode image
	img, _, err := image.Decode(imageData)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate new height maintaining aspect ratio
	aspectRatio := float64(origHeight) / float64(origWidth)
	height := int(aspectRatio * float64(width) * 0.55) // 0.55 compensates for character aspect ratio

	// Convert to ASCII
	var result strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map coordinates to original image
			origX := int(float64(x) * float64(origWidth) / float64(width))
			origY := int(float64(y) * float64(origHeight) / float64(height))

			// Get pixel color and convert to grayscale
			pixel := img.At(origX, origY)
			gray := color.GrayModel.Convert(pixel).(color.Gray)

			// Map grayscale value to ASCII character
			charIndex := int(float64(gray.Y) * float64(len(asciiChars)-1) / 255.0)
			result.WriteByte(asciiChars[charIndex])
		}
		result.WriteByte('\n')
	}

	return result.String(), nil
}

// SaveCaptchaToFile saves captcha image data to a temporary file
func SaveCaptchaToFile(imageData []byte) (string, error) {
	// Generate unique filename
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("captcha_%d.jpg", timestamp)

	// Create temporary file
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create captcha file: %w", err)
	}
	defer file.Close()

	// Write image data
	_, err = file.Write(imageData)
	if err != nil {
		return "", fmt.Errorf("failed to write captcha data: %w", err)
	}

	return filename, nil
}

// OpenImageFile attempts to open an image file with the default system application
func OpenImageFile(filename string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", filename)
	case "darwin":
		cmd = exec.Command("open", filename)
	case "linux":
		cmd = exec.Command("xdg-open", filename)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// DisplayCaptcha displays captcha according to the specified mode and prompts for input
func DisplayCaptcha(imageData []byte, mode CaptchaDisplayMode) (string, error) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("验证码显示")
	fmt.Println(strings.Repeat("=", 60))

	var captchaFile string
	var err error

	// Display ASCII version if requested
	if mode == DisplayASCII || mode == DisplayBoth {
		fmt.Println("\nASCII 艺术版本:")
		fmt.Println(strings.Repeat("-", 40))

		asciiArt, err := ImageToASCII(strings.NewReader(string(imageData)), 60, "standard")
		if err != nil {
			fmt.Printf("ASCII转换失败: %v\n", err)
		} else {
			fmt.Print(asciiArt)
		}
		fmt.Println(strings.Repeat("-", 40))
	}

	// Save to file if requested
	if mode == DisplayFile || mode == DisplayBoth {
		captchaFile, err = SaveCaptchaToFile(imageData)
		if err != nil {
			fmt.Printf("保存验证码文件失败: %v\n", err)
		} else {
			fmt.Printf("\n验证码已保存到文件: %s\n", captchaFile)

			// Try to open the image automatically
			if err := OpenImageFile(captchaFile); err != nil {
				fmt.Printf("无法自动打开图片，请手动查看: %s\n", captchaFile)
			} else {
				fmt.Println("验证码图片已自动打开")
			}
		}
	}

	fmt.Println(strings.Repeat("=", 60))

	// Get user input
	fmt.Print("请输入验证码: ")
	reader := bufio.NewReader(os.Stdin)
	captcha, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read captcha input: %w", err)
	}

	captcha = strings.TrimSpace(captcha)
	if captcha == "" {
		fmt.Println("警告：验证码为空")
	} else {
		fmt.Printf("验证码输入完成: %s\n", captcha)
	}

	// Clean up captcha file immediately
	if captchaFile != "" {
		if err := os.Remove(captchaFile); err != nil {
			fmt.Printf("清理验证码文件失败: %v\n", err)
		} else {
			fmt.Printf("验证码文件已清理: %s\n", captchaFile)
		}
	}

	return captcha, nil
}

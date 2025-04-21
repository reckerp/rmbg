package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/h2non/bimg"
)

const (
	apiURL      = "https://api.remove.bg/v1.0/removebg"
	version     = "1.0.0"
	programName = "rmbg"
)

// Config holds the application configuration
type Config struct {
	Format     string
	Compress   bool
	Quality    int
	ApiKey     string
	InputPath  string
	OutputPath string
}

// ProcessResult represents the result of processing a single image
type ProcessResult struct {
	Filename string
	Success  bool
	Error    error
}

func main() {
	// Parse command-line args manually
	config, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Printf("❌ Error: %s\n", err)
		printUsage()
		os.Exit(1)
	}

	if config.InputPath == "-h" || config.InputPath == "--help" {
		printUsage()
		os.Exit(0)
	}

	// Check for API key
	config.ApiKey = os.Getenv("REMOVE_BG_API_KEY")
	if config.ApiKey == "" {
		fmt.Println("❌ Error: REMOVE_BG_API_KEY environment variable is not set")
		os.Exit(1)
	}

	// Validate input path
	if config.InputPath == "" {
		fmt.Println("❌ Error: Input path is required")
		printUsage()
		os.Exit(1)
	}

	// Process the input path
	fileInfo, err := os.Stat(config.InputPath)
	if err != nil {
		fmt.Printf("❌ Error: %s\n", err)
		os.Exit(1)
	}

	if fileInfo.IsDir() {
		// Process directory
		processDirectory(config)
	} else {
		// Process single file
		outputPath := config.OutputPath
		if outputPath == "" {
			outputPath = generateOutputPath(config.InputPath, config.Format)
		}
		err := processImage(config.InputPath, outputPath, config)
		if err != nil {
			fmt.Printf("❌ Error processing %s: %s\n", config.InputPath, err)
			os.Exit(1)
		}
		fmt.Printf("✅ Successfully processed: %s -> %s\n", config.InputPath, outputPath)
	}
}

// Parse command-line arguments manually
func parseArgs(args []string) (Config, error) {
	config := Config{
		Format:   "png",
		Compress: false,
		Quality:  90,
	}

	var nonFlagArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			// Help flag
			if arg == "-h" || arg == "--help" {
				config.InputPath = "-h"
				return config, nil
			}

			// Format flag
			if arg == "-f" {
				if i+1 >= len(args) {
					return config, fmt.Errorf("missing value for -f flag")
				}

				format := strings.ToLower(args[i+1])
				if format != "png" && format != "webp" {
					fmt.Printf("⚠️ Invalid format: %s. Using default (png).\n", format)
					format = "png"
				}
				config.Format = format
				i++
				continue
			}

			// Compression flag
			if arg == "-c" {
				config.Compress = true

				// Check if next arg is a number (quality parameter)
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					// Try to parse as number
					qualityStr := args[i+1]
					quality, err := strconv.Atoi(qualityStr)

					// If it parses as a number, use it as quality
					if err == nil && quality >= 1 && quality <= 100 {
						config.Quality = quality
						i++
						continue
					}
					// If not a number, then it's likely a file path
				}
				continue
			}

			// Handle combined compression
			if strings.HasPrefix(arg, "-c=") {
				config.Compress = true
				qualityStr := strings.TrimPrefix(arg, "-c=")

				if qualityStr != "" {
					quality, err := strconv.Atoi(qualityStr)
					if err != nil || quality < 1 || quality > 100 {
						return config, fmt.Errorf("invalid quality value: %s", qualityStr)
					}
					config.Quality = quality
				}
				continue
			}

			// Unknown flag
			return config, fmt.Errorf("unknown flag: %s", arg)
		} else {
			// Non-flag argument (input or output path)
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}

	// Set input and output paths
	if len(nonFlagArgs) > 0 {
		config.InputPath = nonFlagArgs[0]
	}

	if len(nonFlagArgs) > 1 {
		config.OutputPath = nonFlagArgs[1]
	}

	return config, nil
}

// Print usage information
func printUsage() {
	fmt.Printf("%s v%s - Background Removal Tool\n\n", programName, version)
	fmt.Println("Usage:")
	fmt.Printf("  %s [options] <input_path> [output_path]\n\n", programName)
	fmt.Println("Options:")
	fmt.Println("  -f <format>")
	fmt.Println("        Output format (png or webp) (default \"png\")")
	fmt.Println("  -c [quality]")
	fmt.Println("        Compress output image. Optionally specify quality (1-100)")
	fmt.Println("        Examples: -c (uses default quality 90), -c=75 (sets quality to 75)")
	fmt.Println("  -h    Display help information")
	fmt.Println("Examples:")
	fmt.Printf("  %s image.jpg                     # Process a single image\n", programName)
	fmt.Printf("  %s -f webp image.jpg             # Process and save as WebP\n", programName)
	fmt.Printf("  %s -c image.jpg                  # Process and compress with default quality\n", programName)
	fmt.Printf("  %s -c 75 image.jpg               # Process and compress with quality 75\n", programName)
	fmt.Printf("  %s -c=75 image.jpg               # Process and compress with quality 75 (alt syntax)\n", programName)
	fmt.Printf("  %s -f webp -c 80 images/         # Process directory as WebP with quality 80\n", programName)
	fmt.Printf("  %s image.jpg custom-output.png   # Process with custom output path\n\n", programName)
	fmt.Println("Notes:")
	fmt.Println("  - The REMOVE_BG_API_KEY environment variable must be set")
	fmt.Println("  - Directory processing creates an output directory with suffix \"-rm\"")
	fmt.Println("  - Supports JPEG, PNG, and WebP input formats")
}

// Generate output path for the processed image
func generateOutputPath(inputPath, format string) string {
	// Extract the base name without extension
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	baseDir := filepath.Dir(inputPath)

	// Use the specified format extension
	return filepath.Join(baseDir, baseName+"-rm."+format)
}

// Process image according to config settings
func optimizeImage(data []byte, config Config) ([]byte, error) {
	// If compression is not requested, return original data
	if !config.Compress {
		return data, nil
	}

	// Load the image using bimg
	image := bimg.NewImage(data)

	// Check if we have a valid image by getting its info
	_, err := image.Size()
	if err != nil {
		return data, err
	}

	// Different optimization strategies based on format
	var options bimg.Options

	switch config.Format {
	case "webp":
		options = bimg.Options{
			Type:          bimg.WEBP,
			Quality:       config.Quality,
			Compression:   6,
			StripMetadata: true,
			Interlace:     false,
		}
	default:
		options = bimg.Options{
			Type:          bimg.PNG,
			Quality:       config.Quality,
			Compression:   6,
			StripMetadata: true,
			Interlace:     false,
		}
	}

	return image.Process(options)
}

// Process a single image
func processImage(inputPath, outputPath string, config Config) error {
	imageData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("image_file", filepath.Base(inputPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, bytes.NewReader(imageData))
	if err != nil {
		return fmt.Errorf("failed to copy image data: %w", err)
	}

	err = writer.WriteField("size", "full")
	if err != nil {
		return fmt.Errorf("failed to write field 'size': %w", err)
	}

	if config.Format == "webp" {
		err = writer.WriteField("format", "webp")
		if err != nil {
			return fmt.Errorf("failed to write field 'format': %w", err)
		}
	}

	writer.Close()

	// Create the HTTP request
	req, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", config.ApiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	outputData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if config.Compress {
		fmt.Printf("🗜️ Compressing image with quality %d...\n", config.Quality)
		originalSize := len(outputData)

		optimizedData, err := optimizeImage(outputData, config)
		if err != nil {
			fmt.Printf("⚠️ Warning: Failed to compress image: %s. Using original output.\n", err)
			// If optimization fails, use the original data
			optimizedData = outputData
		} else {
			optimizedSize := len(optimizedData)

			// Only use the optimized data if it actually reduced the size
			if optimizedSize < originalSize {
				reduction := float64(originalSize-optimizedSize) / float64(originalSize) * 100
				fmt.Printf("📊 Reduced file size by %.1f%% (from %d KB to %d KB) with quality %d\n",
					reduction, originalSize/1024, optimizedSize/1024, config.Quality)

				outputData = optimizedData
			} else {
				fmt.Printf("ℹ️ Compression did not reduce file size. Using original output.\n")
			}
		}
	} else if config.Format == "webp" && !strings.HasSuffix(strings.ToLower(outputPath), ".webp") {
		image := bimg.NewImage(outputData)
		webpData, err := image.Convert(bimg.WEBP)
		if err != nil {
			fmt.Printf("⚠️ Warning: Failed to convert to WebP: %s\n", err)
		} else {
			outputData = webpData
		}
	}

	err = os.WriteFile(outputPath, outputData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// Process an entire directory of images
func processDirectory(config Config) {
	dirPath := config.InputPath

	outputDir := dirPath
	if strings.HasSuffix(dirPath, "/") || strings.HasSuffix(dirPath, "\\") {
		outputDir = strings.TrimSuffix(outputDir, "/")
		outputDir = strings.TrimSuffix(outputDir, "\\")
	}
	outputDir += "-rm"

	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Printf("❌ Error creating output directory: %s\n", err)
		os.Exit(1)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("❌ Error reading directory: %s\n", err)
		os.Exit(1)
	}

	var imagePaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" {
				imagePaths = append(imagePaths, filepath.Join(dirPath, entry.Name()))
			}
		}
	}

	if len(imagePaths) == 0 {
		fmt.Println("No images found in the directory")
		os.Exit(0)
	}

	fmt.Printf("Found %d images to process...\n", len(imagePaths))

	// Process images sequentially
	var failedImages []string
	var successCount int

	for i, imagePath := range imagePaths {
		baseName := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))
		outputPath := filepath.Join(outputDir, baseName+"-rm."+config.Format)

		fmt.Printf("\n🔄 Processing image %d/%d: %s\n", i+1, len(imagePaths), filepath.Base(imagePath))

		err := processImage(imagePath, outputPath, config)
		if err != nil {
			fmt.Printf("❌ Error processing %s: %s\n", filepath.Base(imagePath), err)
			failedImages = append(failedImages, filepath.Base(imagePath))
		} else {
			fmt.Printf("✅ Successfully processed: %s (%d/%d)\n", filepath.Base(imagePath), i+1, len(imagePaths))
			successCount++
		}

		// Add a small delay between API calls to avoid rate limiting
		if i < len(imagePaths)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Print summary
	fmt.Printf("\n✨ Summary: %d/%d images processed successfully\n", successCount, len(imagePaths))
	if len(failedImages) > 0 {
		fmt.Printf("❌ Failed to process %d images:\n", len(failedImages))
		for _, name := range failedImages {
			fmt.Printf("  - %s\n", name)
		}
	}
	fmt.Printf("📁 Output directory: %s\n", outputDir)
}

package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type RedditPost struct {
	Kind string `json:"kind"`
	Data struct {
		Children []struct {
			Data struct {
				Title               string `json:"title"` // Add this line
				CrossPostParentList []struct {
					Media struct {
						RedditVideo struct {
							FallbackURL string `json:"fallback_url"`
						} `json:"reddit_video"`
					} `json:"media"`
				} `json:"crosspost_parent_list"`
				Media struct {
					RedditVideo struct {
						FallbackURL string `json:"fallback_url"`
					} `json:"reddit_video"`
				} `json:"media"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type progressWriter struct {
	total    int64
	written  int64
	writer   io.Writer
	filename string
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err != nil {
		return n, err
	}
	pw.written += int64(n)
	fmt.Printf("\rDownloading: %d%% %s", 100*pw.written/pw.total, pw.filename)
	return n, nil
}

func showBanner() {
	bannerArt := `
######  ####### ######   ######  ######   #####  ######  
##   ## ##      ##   ## ##       ##   ## ##   ## ##   ## 
######  #####   ##   ## ##   ### ######  ####### ######  
##   ## ##      ##   ## ##    ## ##   ## ##   ## ##   ## 
##   ## ####### ######   ######  ##   ## ##   ## ###### 
`
	fmt.Println(bannerArt)
}

func showUsage() {
	usageText := `Usage:

# Download both video and audio.
redgrab REDDIT-VIDEO-URL

# Specify and output directory (default is current directory)
redgrab -o thisdir REDDIT-VIDEO-URL

# Specify a custom User-Agent string (default is "reddit-video-downloader")
redgrab -user-agent "custom-user-agent" REDDIT-VIDEO-URL

# Download video only (no audio)
redgrab -video REDDIT-VIDEO-URL

# Download audio only (no video)
redgrab -audio REDDIT-VIDEO-URL
`
	fmt.Println(usageText)
}

func downloadFile(url string, fileName string, userAgent string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error: failed to download file, status code: %d", resp.StatusCode)
	}

	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	pw := &progressWriter{total: resp.ContentLength, writer: out, filename: fileName}

	_, err = io.Copy(pw, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println()

	return nil
}

func downloadFiles(audioOnly bool, videoOnly bool, audioURL string, videoURL string, postTitle string, userAgent string, outputDir string) (string, string, string, error) {
	videoFile := filepath.Join(outputDir, postTitle+"_video.mp4")
	audioFile := filepath.Join(outputDir, postTitle+"_audio.mp4")
	mergedFile := filepath.Join(outputDir, postTitle+".mp4")

	if audioOnly && videoOnly {
		return "", "", "", fmt.Errorf("Error: Only one flag, either -audio or -video, should be used at a time.")
	}

	if audioOnly {
		// Check if audio file already exists
		if _, err := os.Stat(audioFile); err == nil {
			fmt.Printf("Audio file '%s' already exists. Skipping audio download.\n", audioFile)
		} else {
			err := downloadFile(audioURL, audioFile, userAgent)
			if err != nil {
				return "", "", "", err
			}
		}
	} else if videoOnly {
		// Check if video file already exists
		if _, err := os.Stat(videoFile); err == nil {
			fmt.Printf("Video file '%s' already exists. Skipping video download.\n", videoFile)
		} else {
			err := downloadFile(videoURL, videoFile, userAgent)
			if err != nil {
				return "", "", "", err
			}
		}
	} else {
		// Check if the merged file already exists
		if _, err := os.Stat(mergedFile); err == nil {
			return videoFile, audioFile, mergedFile, nil
		}

		// Check if video file already exists
		if _, err := os.Stat(videoFile); err == nil {
			fmt.Printf("Video file '%s' already exists. Skipping video download.\n", videoFile)
		} else {
			err := downloadFile(videoURL, videoFile, userAgent)
			if err != nil {
				return "", "", "", err
			}
		}

		// Check if audio file already exists
		if _, err := os.Stat(audioFile); err == nil {
			fmt.Printf("Audio file '%s' already exists. Skipping audio download.\n", audioFile)
		} else {
			err := downloadFile(audioURL, audioFile, userAgent)
			if err != nil {
				return "", "", "", err
			}
		}
	}

	return videoFile, audioFile, mergedFile, nil
}

func mergeFiles(audioOnly bool, videoOnly bool, videoFile string, audioFile string, mergedFile string, rawURL string, outputDir string) error {
	// No need to merge audio and video if only one is being downloaded
	if audioOnly || videoOnly || audioFile == "" {
		if videoOnly {
			absVideoFile, err := filepath.Abs(videoFile)
			if err != nil {
				return err
			}
			fmt.Printf("Video saved: %v\n", absVideoFile)
		}

		if audioOnly && audioFile != "" {
			absAudioFile, err := filepath.Abs(audioFile)
			if err != nil {
				return err
			}
			fmt.Printf("Audio saved: %v\n", absAudioFile)
		}
		return nil
	}

	// Get the absolute file path for the merged file
	absMergedFile, err := filepath.Abs(mergedFile)
	if err != nil {
		return err
	}

	// Check if the merged file already exists
	if _, err := os.Stat(mergedFile); err == nil {
		fmt.Printf("Video file already exists: %s\n", absMergedFile)
		return nil
	}

	_, err = exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("Error: ffmpeg is not installed")
	}

	// Create os.File instances for os.DevNull
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return err
	}
	defer devNull.Close()

	// Check if the video file exists
	if _, err := os.Stat(videoFile); os.IsNotExist(err) {
		return fmt.Errorf("Error: video file does not exist: %s", videoFile)
	}

	// Check if the audio file exists
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		return fmt.Errorf("Error: audio file does not exist: %s", audioFile)
	}

	// Get date
	currentTime := time.Now()
	dateString := currentTime.Format("2006-01-02")

	// Meta data string
	metaData := fmt.Sprintf("Downloaded on: %v Source: %s", dateString, rawURL)

	// Merge video and audio
	cmd := exec.Command("ffmpeg", "-y", "-i", videoFile, "-i", audioFile, "-c", "copy", "-metadata", "comment="+metaData, mergedFile)

	// Redirect stdout and stderr to os.DevNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	// Create a buffer to capture stderr output
	var stderr bytes.Buffer

	// Set cmd.Stderr to the buffer
	cmd.Stderr = &stderr

	// Start the ffmpeg command
	err = cmd.Start()
	if err != nil {
		return err
	}

	// Wait for the ffmpeg command to finish
	err = cmd.Wait()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("Error: occurred during merging video and audio: %v\n%s", err, stderr.String())
		} else {
			return err
		}
	} else {
		fmt.Printf("Video (with audio) saved: %v\n", absMergedFile)
	}

	// Delete the audio and video files
	err = os.Remove(videoFile)
	if err != nil {
		return err
	}

	err = os.Remove(audioFile)
	if err != nil {
		return err
	}

	fmt.Println("Standalone audio and video files removed.")

	return nil
}

// sanitizeString for filename
// Remove non standard chars and make OS and terminal friendly
// Prepend date of download and MD5 hash of video URL to filename.
func sanitizeString(str string, videoURL string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}

	sanitizedString := reg.ReplaceAllString(str, "_")

	// Remove trailing underscore if it exists
	if strings.HasSuffix(sanitizedString, "_") {
		sanitizedString = sanitizedString[:len(sanitizedString)-1]
	}

	// Convert to lowercase
	sanitizedString = strings.ToLower(sanitizedString)

	// Restrict length of filename
	if len(sanitizedString) > 60 {
		sanitizedString = sanitizedString[:60]
	}

	// Add current date format YYYYMMDD
	dateString := time.Now().Format("20060102")

	// Calculate MD5 hash of video URL
	hash := md5.Sum([]byte(videoURL))
	hashString := hex.EncodeToString(hash[:])

	// sanitizedString = dateString + "_" + sanitizedString
	sanitizedString = fmt.Sprintf("%v_%v_%s", dateString, hashString, sanitizedString)

	return sanitizedString
}

func convertToBaseURL(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}

	// Decode the URL to handle query parameters
	decodedURL, err := url.QueryUnescape(parsedURL.String())
	if err != nil {
		return "", err
	}

	// Remove the query parameters from the decoded URL
	baseURL := strings.Split(decodedURL, "?")[0]

	return baseURL, nil
}

func parseFlags() (audioOnly bool, videoOnly bool, userAgent string, outputDir string, postURL string, err error) {
	flag.BoolVar(&audioOnly, "audio", false, "Download audio only")
	flag.BoolVar(&videoOnly, "video", false, "Download video only")
	flag.StringVar(&userAgent, "user-agent", "reddit-video-downloader", "Set the user-agent string")
	flag.StringVar(&outputDir, "o", ".", "Specify output directory")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		showUsage()
		err = fmt.Errorf("Target REDDIT-VIDEO-URL required.")
		return
	}

	postURL = args[0]

	return
}

func fetchJSON(client *http.Client, userAgent string, postURL string) ([]RedditPost, error) {
	req, err := http.NewRequest("GET", postURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var posts []RedditPost
	err = json.NewDecoder(resp.Body).Decode(&posts)
	if err != nil {
		return nil, fmt.Errorf("Error: decoding JSON, check URL address: %v\n", err)
	}

	return posts, nil
}

func extractURLs(posts []RedditPost) (string, string, string, error) {
	var videoURL, audioURL, postTitle string

	if len(posts) > 0 && len(posts[0].Data.Children) > 0 {
		post := posts[0].Data.Children[0].Data

		if len(post.Media.RedditVideo.FallbackURL) > 0 {
			videoURL = post.Media.RedditVideo.FallbackURL
		} else if len(post.CrossPostParentList) > 0 {
			videoURL = post.CrossPostParentList[0].Media.RedditVideo.FallbackURL
		}

		if videoURL != "" {
			audioURL = strings.Split(videoURL, "_")[0] + "_audio.mp4"
		}

		postTitle = sanitizeString(posts[0].Data.Children[0].Data.Title, videoURL)
	}

	if videoURL == "" {
		return "", "", "", fmt.Errorf("Error: could not extract video URLs")
	}

	return videoURL, audioURL, postTitle, nil
}

func run() error {
	audioOnly, videoOnly, userAgent, outputDir, postURL, err := parseFlags()
	if err != nil {
		return err
	}

	// Check if output directory exists
	_, err = os.Stat(outputDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("Error: Specified output directory does not exist: %s", outputDir)
	} else if err != nil {
		return fmt.Errorf("Error: Problem with specified output directory: %v", err)
	}

	rawURL := postURL

	postURL, err = convertToBaseURL(postURL)
	if err != nil {
		return fmt.Errorf("Error: processing URL, check URL address: %v\n", err)
	}

	fmt.Printf("Processing URL: %v\n", postURL)

	if !strings.HasSuffix(postURL, "/") {
		postURL += "/"
	}
	postURL += ".json"

	client := &http.Client{}
	posts, err := fetchJSON(client, userAgent, postURL)
	if err != nil {
		return err
	}

	videoURL, audioURL, postTitle, err := extractURLs(posts)
	if err != nil {
		return err
	}

	videoFile, audioFile, mergedFile, err := downloadFiles(audioOnly, videoOnly, audioURL, videoURL, postTitle, userAgent, outputDir)
	if err != nil {
		return err
	}

	err = mergeFiles(audioOnly, videoOnly, videoFile, audioFile, mergedFile, rawURL, outputDir)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	showBanner()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

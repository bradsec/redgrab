# RedGrab

Reddit Video Grabber (redgrab) is a command-line utility for downloading videos from Reddit. Written in Go using standard Go library functions. Program will download from single post video URLs or video URLS from Reddit share-links.

## Prerequisites

- You need to have `ffmpeg` installed on your system to merge the video and audio files (https://ffmpeg.org/).

## Installation

To install redgrab, you need to have Go installed on your system (https://go.dev/doc/install). Once you have Go installed, you can either clone and run from source or download and install with the following command:

```terminal
go install github.com/bradsec/redgrab@latest
```

## Usage 

```terminal
# Download both video and audio.
redgrab REDDIT-VIDEO-URL

# Specify and output directory (default is current directory)
redgrab -o thisdir REDDIT-VIDEO-URL

# Specify a custom User-Agent string (default is: "reddit-video-downloader")
redgrab -user-agent "custom-user-agent" REDDIT-VIDEO-URL

# Download video only (no audio)
redgrab -video REDDIT-VIDEO-URL

# Download audio only (no video)
redgrab -audio REDDIT-VIDEO-URL
```

## File Information

File names are unique with date, MD5 hash of video URL and part of post title for easier identification. Example: `20230516_221fc565d080ee0bef702c3c4bf24a3c_this_post_name.mp4`. FFMPEG will write metadata into the comments including date of download and original URL in the video file (this is only for the complete video not the audio or video only options).

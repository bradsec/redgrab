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
redgrab REDDIT-POST-URL

# Specify and output directory (default is current directory)
redgrab -o thisdir REDDIT-POST-URL

# Specify a custom User-Agent string (default is: "reddit-video-downloader")
redgrab -user-agent "custom-user-agent" REDDIT-POST-URL

# Download video only (no audio)
redgrab -video REDDIT-POST-URL

# Download audio only (no video)
redgrab -audio REDDIT-POST-URL
```

## File Information

File naming consists of the posts date/time YYYYMMDD_HHMM followed but part of title to keep length down. Example: `20230516_2222_this_post_title.mp4`. FFMPEG will write metadata into the comments including date of download and original URL in the video file (this is only for the complete video not the audio or video only options).

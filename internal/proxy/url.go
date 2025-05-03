package proxy

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"imgproxy-proxy/pkg/signing"
)

// ImageOptimizationOptions contains the image optimization parameters.
type ImageOptimizationOptions struct {
	Width   int // Width of the image
	Height  int // Height of the image
	Quality int // Quality of the image (1-100)
}

// GenerateURL constructs an imgproxy URL path based on the provided parameters and configuration.
// It handles URI encoding, extension appending, options inclusion, and signing.
func GenerateURL(uri string, options string, config Config) (string, error) {
	if config.Encode {
		uri = signing.UrlSafeEncode([]byte(uri))
	} else {
		uri = "plain/" + uri
	}

	if options != "" {
		uri = "/" + options + "/" + uri
	} else {
		uri = "/" + uri
	}

	signature, err := signing.Sign(config.Key, config.Salt, uri, config.SignatureSize)
	if err != nil {
		return "", fmt.Errorf("sign error: %w", err)
	}

	finalURL, err := url.JoinPath(config.BaseURL, signature, uri)
	if err != nil {
		return "", fmt.Errorf("url join error: %w", err)
	}

	return finalURL, nil
}

// ParseQueryToOptions converts URL query parameters into ImageOptimizationOptions.
func ParseQueryToOptions(values url.Values) ImageOptimizationOptions {
	var opts ImageOptimizationOptions

	if w := values.Get("w"); w != "" {
		if width, err := strconv.Atoi(w); err == nil {
			opts.Width = width
		}
	}
	if h := values.Get("h"); h != "" {
		if height, err := strconv.Atoi(h); err == nil {
			opts.Height = height
		}
	}
	if q := values.Get("q"); q != "" {
		if quality, err := strconv.Atoi(q); err == nil {
			opts.Quality = quality
		}
	}

	return opts
}

// ParsePathOptions extracts options from the URL path segments.
func ParsePathOptions(pathSegments []string) string {
	var options []string
	validOptions := map[string]bool{
		"w": true, // width
		"h": true, // height
		"q": true, // quality
	}

	for _, segment := range pathSegments {
		if strings.Contains(segment, ":") {
			parts := strings.Split(segment, ":")
			// Only include if it's a valid option type and has a non-empty value
			if len(parts) == 2 && validOptions[parts[0]] && parts[1] != "" {
				// Validate that value is a number
				if _, err := strconv.Atoi(parts[1]); err == nil {
					options = append(options, segment)
				}
			}
		}
	}
	return strings.Join(options, "/")
}

// MergeOptions combines path options with query options, preferring query options.
func MergeOptions(pathOpts string, queryOpts ImageOptimizationOptions) string {
	parts := strings.Split(pathOpts, "/")
	optMap := make(map[string]string)

	// Parse existing path options
	for _, part := range parts {
		if strings.Contains(part, ":") {
			kv := strings.Split(part, ":")
			if len(kv) == 2 {
				optMap[kv[0]] = kv[1]
			}
		}
	}

	// Override with query options
	if queryOpts.Width != 0 {
		optMap["w"] = strconv.Itoa(queryOpts.Width)
	}
	if queryOpts.Height != 0 {
		optMap["h"] = strconv.Itoa(queryOpts.Height)
	}
	if queryOpts.Quality != 0 {
		optMap["q"] = strconv.Itoa(queryOpts.Quality)
	}

	// Build final options string
	var finalOpts []string
	for k, v := range optMap {
		finalOpts = append(finalOpts, k+":"+v)
	}
	return strings.Join(finalOpts, "/")
}

// GetFileExtension extracts the file extension from a URL.
func GetFileExtension(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	path := u.Path
	ext := strings.ToLower(path[strings.LastIndex(path, ".")+1:])
	if strings.Contains(path, ".") && ext != "" {
		return ext
	}
	return ""
}

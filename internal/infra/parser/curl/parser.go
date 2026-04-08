package curl 
import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type ParsedReq struct {
	BaseURL string
	Variables map[string]interface{}
	Features string
	Headers map[string]string
}

func ParseFile(filepath string) (*ParsedReq, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file: %v", err)
	}
	rawStr := string(data)
	result := &ParsedReq{
		Variables: make(map[string]interface{}),
		Headers: make(map[string]string),
	}
	reURL := regexp.MustCompile(`curl\s+'([^']+)'`)
	urlMatch := reURL.FindStringSubmatch(rawStr)
	var rawURL string
	if len(urlMatch) > 1 {
		rawURL = urlMatch[1]
	} else {
		return nil, fmt.Errorf("URL not found in curl command")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL: %v", err)
	}

	result.BaseURL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	q := u.Query()
	varsRaw := q.Get("variables")
	if varsRaw != "" {
		err = json.Unmarshal([]byte(varsRaw), &result.Variables)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse variables JSON: %v", err)
		}
	}

	result.Features = q.Get("features")
	reHeader := regexp.MustCompile(`-H\s+'([^']+)'`)
	headerMatches := reHeader.FindAllStringSubmatch(rawStr, -1)
	for _, match := range headerMatches {
			if len(match) > 1 {
				parts := strings.SplitN(match[1], ": ", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					result.Headers[key] = val
				}
			}
		}	
	return result, nil
}

func PrintResult(p *ParsedReq) {
	fmt.Println("=== HASIL PARSING DYNAMIC CURL ===")
	fmt.Println("\n[BASE URL]:\n", p.BaseURL)
	
	fmt.Println("\n[CURRENT CURSOR]:")
	if cursor, ok := p.Variables["cursor"]; ok {
		fmt.Printf(" -> %v\n", cursor)
	} else {
		fmt.Println(" -> (Tidak ada cursor di query awal)")
	}

	fmt.Println("\n[HEADERS COUNT]:", len(p.Headers), "headers extracted")
	fmt.Println("==================================")
}

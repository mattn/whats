/*
 * whats: A tool to quickly look up something
 *
 * Copyright (c) 2015 Arjun Sreedharan <arjun024@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */
/*
 * whats.go
 * Entry source file
 */

package main

import(
	"os"
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strings"
	"regexp"
	"./whatslib/google"
)

const AUTHOR = "Arjun Sreedharan <arjun024@gmail.com>"
const VERSION = "0.0.1"

const DEBUG = false
const SPACE_URL_ENCODED = "%20"
const REFERER = "http://arjunsreedharan.org"
const GOOGLE_URI = "https://ajax.googleapis.com" +
			"/ajax/services/search/web?v=1.0&q="



func stringify(argv []string) string {
	query := ""
	i := 1
	size := len(argv)
	for i < size {
		query += os.Args[i]
		i++
		if i < size {
			query += SPACE_URL_ENCODED
		}
	}
	return query
}

/* parses json-string and fills the struct */
func parse_json(str []byte, json_ptr *google.GoogleApiDataType) {
	err := json.Unmarshal(str, json_ptr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse json: %s\n",
			err.Error())
		os.Exit(1)
	}
}

func usage() {
	fmt.Printf("%s\n%s%s\n%s%s\n",
		"SYNTAX : whats <SOMETHING>",
		"AUTHOR : ", AUTHOR,
		"VERSION: ", VERSION)
		os.Exit(0)
}

func strip_html(str string) string {
	regexp_html := regexp.MustCompile("<[^>]*>")
	str = regexp_html.ReplaceAllString(str, "")

	replacements := map[string]string {
		"&#8216;" : "'",
		"&#8217;" : "'",
		"&#8220;" : "\"",
		"&#8221;" : "\"",
		"&nbsp;" : " ",
		"&quot;" : "\"",
		"&apos;" : "'",
		"&#34;" : "\"",
		"&#39;" : "'",
		"&amp; " : "& ",
		"&amp;amp; " : "& ",
	}

	for k,v := range replacements {
		str = strings.Replace(str, k, v, -1)
	}

	return str
}

/* From the top 4 results, let me guess which's best */
func guess(r []google.ResultsType) int {
	cues := []string {
		" is a ",
		" are a ",
		" was as ",
		" were a ",
		" defined as ",
		" developed as a ",
	}
	for i, result:= range r {
		if strings.Contains(result.VisibleUrl, "wikipedia.org") {
			return i%3
		}
	}
	for i, result:= range r {
		for _, cue := range cues {
			if strings.Contains(result.Content, cue) {
				return i%3
			}
		}
	}
	return 0
}

func output(g *google.GoogleApiDataType) {
	i := guess((*g).ResponseData.Results)
	content := strip_html((*g).ResponseData.Results[i].Content)
	fmt.Printf("\n%s\n\n", content)
}

func main() {
	var query string
	var gdata google.GoogleApiDataType
	query = GOOGLE_URI + stringify(os.Args)
	if query == GOOGLE_URI {
		usage()
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", query, nil)
	req.Header.Set("Referer", REFERER)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in making HTTP request: %s\n",
			err.Error())
		os.Exit(1)
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in reading HTTP response: %s\n",
			err.Error())
		os.Exit(1)
	}

	parse_json(contents, &gdata)
	output(&gdata);
}

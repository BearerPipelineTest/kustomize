/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsRepoURL(t *testing.T) {

	testcases := []struct {
		input    string
		expected bool
	}{
		{
			input:    "https://github.com/org/repo",
			expected: true,
		},
		{
			input:    "github.com/org/repo",
			expected: true,
		},
		{
			input:    "git@github.com:org/repo",
			expected: true,
		},
		{
			input:    "gh:org/repo",
			expected: true,
		},
		{
			input:    "git::https://gitlab.com/org/repo",
			expected: true,
		},
		{
			input:    "git@gitlab2.sqtools.ru:10022/infra/kubernetes/thanos-base.git?ref=v0.1.0",
			expected: true,
		},
		{
			input:    "git@bitbucket.org:org/repo.git",
			expected: true,
		},
		{
			input:    "git::http://git.example.com/org/repo.git",
			expected: true,
		},
		{
			input:    "git::https://git.example.com/org/repo.git",
			expected: true,
		},
		{
			input:    "ssh://git.example.com:7999/org/repo.git",
			expected: true,
		},
		{
			input:    "/github.com/org/repo",
			expected: false,
		},
		{
			input:    "/abs/path/to/file",
			expected: false,
		},
		{
			input:    "../relative",
			expected: false,
		},
		{
			input:    "foo",
			expected: false,
		},
		{
			input:    ".",
			expected: false,
		},
		{
			input:    "",
			expected: false,
		},
	}
	for _, tc := range testcases {
		actual := IsRepoUrl(tc.input)
		if actual != tc.expected {
			t.Errorf("unexpected error: unexpected result %t for input %s", actual, tc.input)
		}
	}
}

var orgRepos = []string{"someOrg/someRepo", "kubernetes/website"}

var pathNames = []string{"README.md", "foo/krusty.txt", ""}

var hrefArgs = []string{"someBranch", "master", "v0.1.0", ""}

var hostNamesRawAndNormalizedOld = map[string]string{
	"gh:%s":                           "gh:",
	"GH:%s":                           "gh:",
	"git@github.com:%s":               "git@github.com:",
	"gitHub.com/%s":                   "https://github.com/",
	"https://github.com/%s":           "https://github.com/",
	"hTTps://github.com/%s":           "https://github.com/",
	"git::https://gitlab.com/%s":      "https://gitlab.com/",
	"github.com:%s":                   "https://github.com/",
	"git::http://git.example.com/%s":  "http://git.example.com/",
	"git::https://git.example.com/%s": "https://git.example.com/",
	"ssh://git.example.com:7999/%s":   "ssh://git.example.com:7999/",
	"https://git-codecommit.us-east-2.amazonaws.com/%s": "https://git-codecommit.us-east-2.amazonaws.com/",
}

var hostNamesRawAndNormalized = [][]string{
	{"gh:", "gh:"},
	{"GH:", "gh:"},
	{"gitHub.com/", "https://github.com/"},
	{"github.com:", "https://github.com/"},
	{"http://github.com/", "https://github.com/"},
	{"https://github.com/", "https://github.com/"},
	{"hTTps://github.com/", "https://github.com/"},
	{"https://git-codecommit.us-east-2.amazonaws.com/", "https://git-codecommit.us-east-2.amazonaws.com/"},
	{"https://fabrikops2.visualstudio.com/", "https://fabrikops2.visualstudio.com/"},
	{"ssh://git.example.com:7999/", "ssh://git.example.com:7999/"},
	{"git::https://gitlab.com/", "https://gitlab.com/"},
	{"git::http://git.example.com/", "http://git.example.com/"},
	{"git::https://git.example.com/", "https://git.example.com/"},
	{"git@github.com:", "git@github.com:"},
	{"git@github.com/", "git@github.com:"},
	{"git@gitlab2.sqtools.ru:10022/", "git@gitlab2.sqtools.ru:10022/"},
}

func makeUrlOld(hostFmt, orgRepo, path, href string) string {
	if len(path) > 0 {
		orgRepo = filepath.Join(orgRepo, path)
	}
	url := fmt.Sprintf(hostFmt, orgRepo)
	if href != "" {
		url += refQuery + href
	}
	return url
}

func makeUrl(hostFmt, orgRepo, path, href string) string {
	if len(path) > 0 {
		orgRepo = filepath.Join(orgRepo, path)
	}
	url := hostFmt + orgRepo
	if href != "" {
		url += refQuery + href
	}
	return url
}

func TestNewRepoSpecFromUrl(t *testing.T) {
	var bad [][]string
	for _, tuple := range hostNamesRawAndNormalized {
		hostRaw := tuple[0]
		hostSpec := tuple[1]
		for _, orgRepo := range orgRepos {
			for _, pathName := range pathNames {
				for _, hrefArg := range hrefArgs {
					uri := makeUrl(hostRaw, orgRepo, pathName, hrefArg)
					rs, err := NewRepoSpecFromUrl(uri)
					if err != nil {
						t.Errorf("problem %v", err)
					}
					if rs.host != hostSpec {
						bad = append(bad, []string{"host", uri, rs.host, hostSpec})
					}
					if rs.orgRepo != orgRepo {
						bad = append(bad, []string{"orgRepo", uri, rs.orgRepo, orgRepo})
					}
					if rs.path != pathName {
						bad = append(bad, []string{"path", uri, rs.path, pathName})
					}
					if rs.ref != hrefArg {
						bad = append(bad, []string{"ref", uri, rs.ref, hrefArg})
					}
				}
			}
		}
	}
	if len(bad) > 0 {
		for _, tuple := range bad {
			fmt.Printf("\n"+
				"     from uri: %s\n"+
				"  actual %4s: %s\n"+
				"expected %4s: %s\n",
				tuple[1], tuple[0], tuple[2], tuple[0], tuple[3])
		}
		t.Fail()
	}
}

var badData = [][]string{
	{"/tmp", "uri looks like abs path"},
	{"iauhsdiuashduas", "url lacks orgRepo"},
	{"htxxxtp://github.com/", "url lacks host"},
	{"ssh://git.example.com", "url lacks orgRepo"},
	{"git::___", "url lacks orgRepo"},
}

func TestNewRepoSpecFromUrlErrors(t *testing.T) {
	for _, tuple := range badData {
		_, err := NewRepoSpecFromUrl(tuple[0])
		if err == nil {
			t.Error("expected error")
		}
		if !strings.Contains(err.Error(), tuple[1]) {
			t.Errorf("unexpected error: %s", err)
		}
	}
}

func TestParseGithubUrl(t *testing.T) {
	for hostRawFmt, hostSpec := range hostNamesRawAndNormalizedOld {
		for _, orgRepo := range orgRepos {
			for _, pathName := range pathNames {
				for _, hrefArg := range hrefArgs {
					input := makeUrlOld(hostRawFmt, orgRepo, pathName, hrefArg)
					if !IsRepoUrl(input) {
						t.Errorf("Should smell like github arg: %s\n", input)
					}
					host, repo, path, gitRef := parseGithubUrl(input)
					if host != hostSpec {
						t.Errorf("\n"+
							"         from %s\n"+
							"  actual host %s\n"+
							"expected host %s\n", input, host, hostSpec)
					}
					if repo != orgRepo {
						t.Errorf("\n"+
							"         from %s\n"+
							"  actual Repo %s\n"+
							"expected Repo %s\n", input, repo, orgRepo)
					}
					if path != pathName {
						t.Errorf("\n"+
							"         from %s\n"+
							"  actual Path %s\n"+
							"expected Path %s\n", input, path, pathName)
					}
					if gitRef != hrefArg {
						t.Errorf("\n"+
							"         from %s\n"+
							"  actual Href %s\n"+
							"expected Href %s\n", input, gitRef, hrefArg)
					}
				}
			}
		}
	}
}

func TestNewRepoSpecFromUrlOld(t *testing.T) {
	testcases := []struct {
		input string
		repo  string
		path  string
		ref   string
	}{
		{
			input: "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo/somedir",
			repo:  "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo",
			path:  "somedir",
			ref:   "",
		},
		{
			input: "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo/somedir?ref=testbranch",
			repo:  "https://git-codecommit.us-east-2.amazonaws.com/someorg/somerepo",
			path:  "somedir",
			ref:   "testbranch",
		},
		{
			input: "https://fabrikops2.visualstudio.com/someorg/somerepo?ref=master",
			repo:  "https://fabrikops2.visualstudio.com/someorg/somerepo",
			path:  "",
			ref:   "master",
		},
		{
			input: "http://github.com/someorg/somerepo/somedir",
			repo:  "https://github.com/someorg/somerepo.git",
			path:  "somedir",
			ref:   "",
		},
		{
			input: "git@github.com:someorg/somerepo/somedir",
			repo:  "git@github.com:someorg/somerepo.git",
			path:  "somedir",
			ref:   "",
		},
		{
			input: "git@gitlab2.sqtools.ru:10022/infra/kubernetes/thanos-base.git?ref=v0.1.0",
			repo:  "git@gitlab2.sqtools.ru:10022/infra/kubernetes/thanos-base.git",
			path:  "",
			ref:   "v0.1.0",
		},
	}
	for _, testcase := range testcases {
		rs, err := NewRepoSpecFromUrl(testcase.input)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if rs.CloneSpec() != testcase.repo {
			t.Errorf("CloneSpec expected to be %v, but got %v on %s",
				testcase.repo, rs.CloneSpec(), testcase.input)
		}
		if rs.path != testcase.path {
			t.Errorf("path expected to be %v, but got %v on %s",
				testcase.path, rs.path, testcase.input)
		}
		if rs.ref != testcase.ref {
			t.Errorf("ref expected to be %v, but got %v on %s",
				testcase.ref, rs.ref, testcase.input)
		}
	}
}

func TestIsAzureHost(t *testing.T) {
	testcases := []struct {
		input  string
		expect bool
	}{
		{
			input:  "https://git-codecommit.us-east-2.amazonaws.com",
			expect: false,
		},
		{
			input:  "ssh://git-codecommit.us-east-2.amazonaws.com",
			expect: false,
		},
		{
			input:  "https://fabrikops2.visualstudio.com/",
			expect: true,
		},
		{
			input:  "https://dev.azure.com/myorg/myproject/",
			expect: true,
		},
	}
	for _, testcase := range testcases {
		actual := isAzureHost(testcase.input)
		if actual != testcase.expect {
			t.Errorf("IsAzureHost: expected %v, but got %v on %s", testcase.expect, actual, testcase.input)
		}
	}
}

func TestIsAWSHost(t *testing.T) {
	testcases := []struct {
		input  string
		expect bool
	}{
		{
			input:  "https://git-codecommit.us-east-2.amazonaws.com",
			expect: true,
		},
		{
			input:  "ssh://git-codecommit.us-east-2.amazonaws.com",
			expect: true,
		},
		{
			input:  "git@github.com:",
			expect: false,
		},
		{
			input:  "http://github.com/",
			expect: false,
		},
	}
	for _, testcase := range testcases {
		actual := isAWSHost(testcase.input)
		if actual != testcase.expect {
			t.Errorf("IsAWSHost: expected %v, but got %v on %s", testcase.expect, actual, testcase.input)
		}
	}
}

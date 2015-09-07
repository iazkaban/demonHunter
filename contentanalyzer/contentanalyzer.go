package contentanalyzer

import (
	"bytes"
	"container/list"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/iazkaban/demonHunter/config"
	"github.com/iazkaban/demonHunter/login"
)

var urllist *list.List
var urlmap map[[16]byte]bool
var unrulyList []*regexp.Regexp
var setUrlSwitch bool
var lastSetUrlTime time.Time
var lastGetUrlTime time.Time
var maxSetUrlTimeTimeout time.Duration
var maxGetUrlTimeTimeout time.Duration

type Page struct {
	Url      string
	isTarget bool
	Body     []byte
	Links    [][]byte
}

func init() {
	urllist = list.New()
	urllist.Init()
	urlmap = make(map[[16]byte]bool, 1024)

	setUrlSwitch = true
	lastSetUrlTime = time.Now()
	lastGetUrlTime = time.Now()
	maxSetUrlTimeTimeout, _ = time.ParseDuration("5m")
	maxGetUrlTimeTimeout, _ = time.ParseDuration("5m")
}

func SetUrl(url string) error {
	if !setUrlSwitch {
		return errors.New("Set Url Switch Is False.")
	}
	if time.Now().Sub(lastSetUrlTime) > maxSetUrlTimeTimeout {
		setUrlSwitch = false
		return errors.New("Set Url Switch Time Out.")
	}
	url_md5 := md5.Sum([]byte(url))
	if _, ok := urlmap[url_md5]; ok {
		return nil
	}
	urlmap[url_md5] = true
	urllist.PushBack(url)
	return nil
}

func Run() {
	if len(config.Config.Server.UrlUnruly) > 0 {
		for _, value := range config.Config.Server.UrlUnruly {
			unruly := regexp.MustCompile(value)
			unrulyList = append(unrulyList, unruly)
		}
	}
	for {
		if time.Now().Sub(lastGetUrlTime) > maxGetUrlTimeTimeout {
			break
		}
		e := urllist.Front()
		if e == nil {
			continue
		}
		urllist.Remove(e)
		lastGetUrlTime = time.Now()
		if len(unrulyList) > 0 {
			tmp := false
			for _, reg := range unrulyList {
				if reg.MatchString(e.Value.(string)) {
					tmp = true
					break
				}
			}
			if tmp == true {
				continue
			}
		}
		if !checkLimitUrl(e.Value.(string)) {
			continue
		}
		switch e.Value.(type) {
		case string:
			url, ok := e.Value.(string)
			if ok {
				Analyzer(&Page{Url: url})
			}
		default:
		}
		e.Next()
	}
}

func Analyzer(page *Page) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", page.Url, strings.NewReader(""))
	if err != nil {
		return err
	}
	req.Header = login.Header
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	full_body := make([]byte, 1)
	body := make([]byte, 1)
	for {
		_, err = response.Body.Read(body)
		full_body = append(full_body, body...)
		if err == io.EOF {
			break
		}
	}
	page.Body = full_body
	page.Links = GetUrls(full_body)
	fmt.Println(page.Url)
	for _, v := range config.Config.Server.UrlRules {
		reg := regexp.MustCompile(v)
		if reg.FindString(page.Url) != "" {
			fmt.Println(string(full_body))
			err = saveFile(full_body)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	for _, v := range page.Links {
		tmp_url, err := checkUrl(page.Url, string(v))
		if err != nil {
			continue
		}
		SetUrl(tmp_url)
	}
	return nil
}

func GetUrls(body []byte) [][]byte {
	reg := regexp.MustCompile("<a .+>")
	urlreg := regexp.MustCompile(`href=['"].+?['"]`)
	a_s := reg.FindAll(body, -1)
	if a_s == nil {
		return nil
	}
	for i, _ := range a_s {
		a_s[i] = urlreg.Find(a_s[i])
		if a_s[i] != nil {
			a_s[i] = a_s[i][:len(a_s[i])]
			a_s[i] = bytes.Replace(a_s[i], []byte("href="), []byte(""), -1)
			a_s[i] = bytes.Trim(a_s[i], `"'`)
		}
	}
	return a_s
}

func checkUrl(parentUrl, cururl string) (string, error) {
	//检查是否是#开头得内连链接
	if index := strings.IndexByte(cururl, '#'); index == 0 {
		return "", errors.New("# on Head.")
	}
	//检查是否完整的http(s)绝对链接
	reg := regexp.MustCompile(`^https?:\/\/`)
	rs := reg.FindString(cururl)
	if rs != "" {
		return cururl, nil
	}
	//解析父级页面链接
	u, err := url.Parse(parentUrl)
	if err != nil {
		return "", err
	}
	//检查是否是/开头得相对链接
	if index := strings.IndexByte(cururl, '/'); index == 0 {
		return u.Scheme + "://" + u.Host + cururl, nil
	}
	//检查是否是数字和字母开头得相对连接
	reg = regexp.MustCompile(`^[a-zA-Z0-9]`)
	rs = reg.FindString(cururl)
	if rs != "" {
		return u.Scheme + "://" + u.Host + cururl, nil
	}
	return "", errors.New("Parse error")
}

func checkLimitUrl(iurl string) bool {
	url_obj, err := url.Parse(iurl)
	if err != nil {
		return false
	}
	for _, v := range config.Config.Server.LimitHost {
		if url_obj.Host == v {
			return true
		}
	}
	return false
}

func saveFile(body []byte) error {
	static_resource := [][]byte{
		[]byte("/s/zh/2154/29/131/_/download/superbatch/css/batch.css"),
		[]byte("/s/zh/2154/29/131/_/download/superbatch/css/batch.css?media=Println"),
		[]byte("/s/zh/2154/29/1.0/_/download/resources/confluence.web.resources:aui-forms/confluence-forms.css"),
		[]byte("/s/zh/2154/29/0.8/_/download/batch/com.atlassian.plugins.shortcuts.atlassian-shortcuts-module:shortcuts/com.atlassian.plugins.shortcuts.atlassian-shortcuts-module:shortcuts.css"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.keyboardshortcuts:confluence-keyboard-shortcuts/com.atlassian.confluence.keyboardshortcuts:confluence-keyboard-shortcuts.css"),
		[]byte("/s/zh/2154/29/29/_/styles/combined.css"),
		[]byte("/s/zh/2154/29/131/_/download/superbatch/js/batch.js"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/confluence.web.resources:login/confluence.web.resources:login.js"),
		[]byte("/s/zh/2154/29/1.9/_/download/batch/com.atlassian.confluence.plugins.doctheme:splitter/com.atlassian.confluence.plugins.doctheme:splitter.js"),
		[]byte("/s/zh/2154/29/0.8/_/download/batch/com.atlassian.plugins.shortcuts.atlassian-shortcuts-module:shortcuts/com.atlassian.plugins.shortcuts.atlassian-shortcuts-module:shortcuts.js"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.keyboardshortcuts:confluence-keyboard-shortcuts/com.atlassian.confluence.keyboardshortcuts:confluence-keyboard-shortcuts.js"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/legacy.confluence.web.resources:prototype/legacy.confluence.web.resources:prototype.js"),
		[]byte("/s/zh/2154/29/1.16/_/download/batch/confluence.macros.advanced:fancy-box/confluence.macros.advanced:fancy-box.js"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/confluence.web.resources:page-editor/confluence.web.resources:page-editor.js"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/confluence.web.resources:create-page-editor/confluence.web.resources:create-page-editor.js"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.tinymceplugin:editor-resources/com.atlassian.confluence.tinymceplugin:editor-resources.js"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/resources/com.atlassian.confluence.plugins.linkbrowser:linkbrowser-resources/war/linkbrowser/linkbrowser.nocache.js"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.plugins.linkbrowser:linkbrowser-resources/com.atlassian.confluence.plugins.linkbrowser:linkbrowser-resources.js"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/confluence-draft-changes:draft-changes/confluence-draft-changes:draft-changes.js"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.tinymceplugin:editor-autocomplete-resources/com.atlassian.confluence.tinymceplugin:editor-autocomplete-resources.js"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.tinymceplugin:editor-autocomplete-links/com.atlassian.confluence.tinymceplugin:editor-autocomplete-links.js"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.tinymceplugin:editor-autocomplete-media/com.atlassian.confluence.tinymceplugin:editor-autocomplete-media.js"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.tinymceplugin:editor-autocomplete-macros/com.atlassian.confluence.tinymceplugin:editor-autocomplete-macros.js"),
		[]byte("/s/zh/2154/29/3.3/_/download/batch/com.atlassian.applinks.applinks-plugin:applinks-util-js/com.atlassian.applinks.applinks-plugin:applinks-util-js.js"),
		[]byte("/s/zh/2154/29/3.3/_/download/batch/com.atlassian.applinks.applinks-plugin:applinks-oauth-ui/com.atlassian.applinks.applinks-plugin:applinks-oauth-ui.js"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/com.atlassian.confluence.plugins.jira.jira-connector:proxy-js/com.atlassian.confluence.plugins.jira.jira-connector:proxy-js.js"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/com.atlassian.confluence.plugins.jira.jira-connector:dialogsJs/com.atlassian.confluence.plugins.jira.jira-connector:dialogsJs.js"),
		[]byte("/s/zh/2154/29/1.15/_/download/batch/com.atlassian.confluence.extra.officeconnector:macro-browser-smart-fields/com.atlassian.confluence.extra.officeconnector:macro-browser-smart-fields.js"),
		[]byte("/s/zh/2154/29/2.0.8/_/download/batch/com.atlassian.gadgets.embedded:gadget-core-resources/com.atlassian.gadgets.embedded:gadget-core-resources.js"),
		[]byte("/s/zh/2154/29/1.1.5/_/download/batch/com.atlassian.confluence.plugins.gadgets:macro-browser-for-gadgetsplugin/com.atlassian.confluence.plugins.gadgets:macro-browser-for-gadgetsplugin.js"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/confluence.macros.core:macro-browser-smart-fields/confluence.macros.core:macro-browser-smart-fields.js"),
		[]byte("/s/zh/2154/29/1.6/_/download/batch/com.atlassian.confluence.plugins.share-page:mail-page-resources/com.atlassian.confluence.plugins.share-page:mail-page-resources.js"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/confluence.web.resources:help-content-resources/confluence.web.resources:help-content-resources.js"),
		[]byte("/s/zh/2154/29/1.16/_/download/batch/confluence.macros.advanced:fancy-box/confluence.macros.advanced:fancy-box.css"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/confluence.web.resources:page-editor/confluence.web.resources:page-editor.css"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/confluence.web.resources:page-editor/confluence.web.resources:page-editor.css?ieonly=true"),
		[]byte("/s/zh/2154/29/3.1/_/download/batch/confluence.extra.jira:macro-browser-resources/confluence.extra.jira:macro-browser-resources.css"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.tinymceplugin:editor-resources/com.atlassian.confluence.tinymceplugin:editor-resources.css"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.plugins.linkbrowser:linkbrowser-resources/com.atlassian.confluence.plugins.linkbrowser:linkbrowser-resources.css"),
		[]byte("/s/zh/2154/29/3.5.3/_/download/batch/com.atlassian.confluence.plugins.linkbrowser:linkbrowser-resources/com.atlassian.confluence.plugins.linkbrowser:linkbrowser-resources.css?ieonly=true"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/confluence-draft-changes:draft-changes/confluence-draft-changes:draft-changes.css"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/com.atlassian.confluence.plugins.jira.jira-connector:dialogsJs/com.atlassian.confluence.plugins.jira.jira-connector:dialogsJs.css"),
		[]byte("/s/zh/2154/29/1.0/_/download/batch/confluence.web.resources:aui-forms/confluence.web.resources:aui-forms.css?ieonly=true"),
		[]byte("/s/zh/2154/29/1.6/_/download/batch/com.atlassian.confluence.plugins.share-page:mail-page-resources/com.atlassian.confluence.plugins.share-page:mail-page-resources.css"),
	}
	reg := regexp.MustCompile(`<title>.+</title>`)
	title := reg.Find(body)
	title = bytes.Replace(bytes.Replace(title, []byte("<title>"), []byte{}, -1), []byte("</title>"), []byte{}, -1)
	file1, err := os.Create("/Users/Logan/Developer/gopath/src/github.com/iazkaban/demonHunter/result/" + string(title) + ".src.html")
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file1.Close()
	file1.Write(body)
	domain_len := len([]byte("http://wiki.elex-tech.com"))
	for _, value := range static_resource {
		source_len := len(value)
		result := make([]byte, domain_len+source_len)
		copy(result[0:domain_len], []byte("http://wiki.elex-tech.com"))
		copy(result[domain_len:domain_len+source_len], value)
		body = bytes.Replace(body, value, result, -1)
	}

	file, err := os.Create("/Users/Logan/Developer/gopath/src/github.com/iazkaban/demonHunter/result/" + string(title) + ".html")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return err
	}
	defer file.Close()
	_, err = file.Write(body)
	os.Exit(1)
	return nil
}

package net

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/jasonzyt/genshincard/config"
)

type PlayerProfileApiUrls struct {
	PlayerIndexUrl       string
	PlayerCharacterUrl   string
	PlayerSprialAbyssUrl string
}

const (
	MiYouSheApi = 0
	HoYoLabApi  = 1

	ChinaOfficialRegion = "cn_gf01"
	ChinaBilibiliRegion = "cn_qd01"
	NorthAmericaRegion  = "os_usa"
	EuropeRegion        = "os_euro"
	AsiaRegion          = "os_asia"
	SarRegion           = "os_cht" // Hong Kong, Macau, Taiwan
)

var (
	apiUrls = map[int]PlayerProfileApiUrls{
		MiYouSheApi: {
			PlayerIndexUrl:       "https://api-takumi-record.mihoyo.com/game_record/app/genshin/api/index",
			PlayerCharacterUrl:   "https://api-takumi-record.mihoyo.com/game_record/app/genshin/api/character",
			PlayerSprialAbyssUrl: "https://api-takumi-record.mihoyo.com/game_record/app/genshin/api/spiralAbyss",
		},
		HoYoLabApi: {
			PlayerIndexUrl:       "https://bbs-api-os.hoyolab.com/game_record/genshin/api/index",
			PlayerCharacterUrl:   "", // Not supported
			PlayerSprialAbyssUrl: "https://bbs-api-os.hoyolab.com/game_record/genshin/api/spiralAbyss",
		},
	}
)

func httpGet(apiType int, region string, uid string) (*http.Response, error) {
	query := fmt.Sprintf("role_id=%s&server=%s", uid, region)
	url := apiUrls[apiType].PlayerIndexUrl + "?" + query
	salt := "xV8v4Qu54lUKrEYFZkJhB8cuOh9Asafs"
	// Current unix timestamp in seconds
	t := time.Now().UnixNano() / 1e9
	// Generate random number bwtween 100 000 and 200 000. If it is exactly 100000, assign 642367
	r := rand.Intn(100000) + 100000
	if r == 100000 {
		r = 642367
	}

	text := fmt.Sprintf("salt=%s&t=%d&r=%d&b=&q=%s", salt, t, r, query)
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(text))
	cipherStr := md5Ctx.Sum(nil)
	sign := fmt.Sprintf("%x", cipherStr)
	final := fmt.Sprintf("%d,%d,%s", t, r, sign)
	// fmt.Printf("[DEBUG] final DS: %s\n", final)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Referer", "https://webstatic.mihoyo.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 13; M2101K9C Build/TKQ1.220829.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/108.0.5359.128 Mobile Safari/537.36 miHoYoBBS/2.44.1")
	req.Header.Set("X-Requested-With", "com.mihoyo.hyperion")
	req.Header.Set("DS", final)
	req.Header.Set("Origin", "https://api-takumi-record.mihoyo.com")
	req.Header.Set("Host", "api-takumi-record.mihoyo.com")
	req.Header.Set("x-rpc-app_version", "2.44.1")
	req.Header.Set("x-rpc-client_type", "5")
	req.Header.Set("Cookie", config.GlobalConfig.CookieSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http get %s failed, status: %s", url, resp.Status)
	}

	return resp, nil
}

func QueryPlayerProfile(apiType int, region string, uid string) (*PlayerProfile, error) {
	resp, err := httpGet(apiType, region, uid)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var deserializedResp PlayerIndexResponse
	err = json.NewDecoder(resp.Body).Decode(&deserializedResp)
	if err != nil || deserializedResp.ReturnCode != 0 {
		return nil, err
	}
	deserializedResp.Data.RoleId = uid
	return &deserializedResp.Data, nil
}

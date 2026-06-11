package utils_test

import (
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/synctv-org/synctv/utils"
)

func TestGetPageItemsRange(t *testing.T) {
	tests := []struct {
		name      string
		total     int
		page      int
		pageSize  int
		wantStart int
		wantEnd   int
	}{
		{"first page", 10, 1, 5, 0, 5},
		{"second page", 10, 2, 5, 5, 10},
		{"page beyond total", 10, 3, 5, 10, 10},
		{"partial last page", 7, 2, 5, 5, 7},
		{"zero page", 10, 0, 5, 0, 0},
		{"zero pageSize", 10, 1, 0, 0, 0},
		{"negative page", 10, -1, 5, 0, 0},
		{"empty total", 0, 1, 5, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := utils.GetPageItemsRange(tt.total, tt.page, tt.pageSize)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("GetPageItemsRange(%d,%d,%d) = (%d,%d), want (%d,%d)",
					tt.total, tt.page, tt.pageSize, start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestIndexAndIn(t *testing.T) {
	items := []string{"a", "b", "c"}
	if got := utils.Index(items, "b"); got != 1 {
		t.Errorf("Index = %d, want 1", got)
	}
	if got := utils.Index(items, "z"); got != -1 {
		t.Errorf("Index missing = %d, want -1", got)
	}
	if !utils.In(items, "c") {
		t.Error("In should find existing item")
	}
	if utils.In(items, "z") {
		t.Error("In should not find missing item")
	}
	if utils.In([]int{}, 1) {
		t.Error("In on empty slice should be false")
	}
}

func TestSplitVersion(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []int
		wantErr bool
	}{
		{"simple", "1.2.3", []int{1, 2, 3}, false},
		{"single", "10", []int{10}, false},
		{"invalid", "1.x.3", nil, true},
		{"empty segment", "1..3", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.SplitVersion(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SplitVersion err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitVersion = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompVersion(t *testing.T) {
	tests := []struct {
		name    string
		v1, v2  string
		want    int
		wantErr bool
	}{
		{"equal", "v1.0.0", "v1.0.0", utils.VersionEqual, false},
		{"greater patch", "v1.0.1", "v1.0.0", utils.VersionGreater, false},
		{"less minor", "v1.0.0", "v1.1.0", utils.VersionLess, false},
		{"release gt prerelease", "v1.0.0", "v1.0.0-beta.1", utils.VersionGreater, false},
		{"prerelease lt release", "v1.0.0-alpha.1", "v1.0.0", utils.VersionLess, false},
		{"length mismatch", "v1.0", "v1.0.0", utils.VersionEqual, true},
		{"invalid", "vx.0.0", "v1.0.0", utils.VersionEqual, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.CompVersion(tt.v1, tt.v2)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CompVersion err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CompVersion(%q,%q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestLIKE(t *testing.T) {
	if got := utils.LIKE("foo"); got != "%foo%" {
		t.Errorf("LIKE = %q, want %%foo%%", got)
	}
	if got := utils.LIKE(""); got != "%%" {
		t.Errorf("LIKE empty = %q, want %%%%", got)
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"video.mp4", "mp4"},
		{"archive.tar.gz", "gz"},
		{"noext", ""},
		{".hidden", "hidden"},
		{"path/to/file.M3U8", "M3U8"},
	}
	for _, tt := range tests {
		if got := utils.GetFileExtension(tt.in); got != tt.want {
			t.Errorf("GetFileExtension(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGetURLExtension(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"https://host/path/movie.mp4", "mp4"},
		{"https://host/stream.m3u8?token=abc", "m3u8"},
		{"https://host/play?file=video.flv", "flv"},
		{"https://host/nopath", ""},
	}
	for _, tt := range tests {
		if got := utils.GetURLExtension(tt.in); got != tt.want {
			t.Errorf("GetURLExtension(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestIsM3u8Url(t *testing.T) {
	if !utils.IsM3u8Url("https://host/live.m3u8") {
		t.Error("expected m3u8 url to be detected")
	}
	if !utils.IsM3u8Url("https://host/live.m3u") {
		t.Error("expected m3u url to be detected")
	}
	if utils.IsM3u8Url("https://host/movie.mp4") {
		t.Error("mp4 should not be detected as m3u8")
	}
}

func TestHTTPCookieMapRoundTrip(t *testing.T) {
	cookies := []*http.Cookie{
		{Name: "session", Value: "abc"},
		{Name: "token", Value: "xyz"},
	}
	m := utils.HTTPCookieToMap(cookies)
	if len(m) != 2 || m["session"] != "abc" || m["token"] != "xyz" {
		t.Fatalf("HTTPCookieToMap = %v", m)
	}

	back := utils.MapToHTTPCookie(m)
	if len(back) != 2 {
		t.Fatalf("MapToHTTPCookie len = %d, want 2", len(back))
	}
	got := utils.HTTPCookieToMap(back)
	if !reflect.DeepEqual(got, m) {
		t.Errorf("round-trip mismatch: %v != %v", got, m)
	}
}

func TestSortUUIDWithUUID(t *testing.T) {
	id := uuid.MustParse("12345678-1234-1234-1234-1234567890ab")
	got := utils.SortUUIDWithUUID(id)
	want := "123456781234123412341234567890ab"
	if got != want {
		t.Errorf("SortUUIDWithUUID = %q, want %q", got, want)
	}
	if len(utils.SortUUID()) != 32 {
		t.Errorf("SortUUID length = %d, want 32", len(utils.SortUUID()))
	}
}

func TestIsLocalIPSSRFGuard(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{"loopback v4", "127.0.0.1", true},
		{"loopback v4 non-one", "127.0.0.2", true},
		{"loopback v4 with port", "127.0.0.1:8080", true},
		{"loopback v6", "::1", true},
		{"loopback v6 bracketed", "[::1]:8080", true},
		{"private 10/8", "10.1.2.3", true},
		{"private 172.16/12", "172.16.5.4", true},
		{"private 192.168/16", "192.168.1.1", true},
		{"link-local metadata", "169.254.169.254", true},
		{"unspecified v4", "0.0.0.0", true},
		{"unique local v6", "fd00::1", true},
		{"public v4", "8.8.8.8", false},
		{"public v6", "2001:4860:4860::8888", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsLocalIP(tt.addr); got != tt.want {
				t.Errorf("IsLocalIP(%q) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}

func TestOnceDoAndReset(t *testing.T) {
	var o utils.Once
	var count int32
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.Do(func() { atomic.AddInt32(&count, 1) })
		}()
	}
	wg.Wait()
	if atomic.LoadInt32(&count) != 1 {
		t.Fatalf("Do ran %d times, want 1", count)
	}

	o.Reset()
	o.Do(func() { atomic.AddInt32(&count, 1) })
	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("after Reset Do ran, count = %d, want 2", count)
	}
}

package dnsforward

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsfilter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openCases(t *testing.T, casesFileName string, cases interface{}) {
	t.Helper()

	var f *os.File
	f, err := os.Open(filepath.Join("testdata", casesFileName))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	err = json.NewDecoder(f).Decode(cases)
	require.NoError(t, err)
}

func TestDNSForwardHTTTP_handleGetConfig(t *testing.T) {
	filterConf := &dnsfilter.Config{
		SafeBrowsingEnabled:   true,
		SafeBrowsingCacheSize: 1000,
		SafeSearchEnabled:     true,
		SafeSearchCacheSize:   1000,
		ParentalCacheSize:     1000,
		CacheTime:             30,
	}
	forwardConf := ServerConfig{
		UDPListenAddrs: []*net.UDPAddr{},
		TCPListenAddrs: []*net.TCPAddr{},
		FilteringConfig: FilteringConfig{
			ProtectionEnabled: true,
			UpstreamDNS:       []string{"8.8.8.8:53", "8.8.4.4:53"},
		},
		ConfigModified: func() {},
	}
	s := createTestServer(t, filterConf, forwardConf)
	require.Nil(t, s.Start())
	t.Cleanup(func() {
		require.Nil(t, s.Stop())
	})

	defaultConf := s.conf

	w := httptest.NewRecorder()

	const casesFileName = "handleGetConfig_cases.json"
	testCases := []struct {
		Name string `json:"name"`
		conf func() ServerConfig
		Want string `json:"want"`
	}{{
		conf: func() ServerConfig {
			return defaultConf
		},
	}, {
		conf: func() ServerConfig {
			conf := defaultConf
			conf.FastestAddr = true

			return conf
		},
	}, {
		conf: func() ServerConfig {
			conf := defaultConf
			conf.AllServers = true

			return conf
		},
	}}
	openCases(t, casesFileName, &testCases)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Cleanup(w.Body.Reset)

			s.conf = tc.conf()
			s.handleGetConfig(w, nil)

			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, tc.Want, w.Body.String())
		})
	}
}

func TestDNSForwardHTTTP_handleSetConfig(t *testing.T) {
	filterConf := &dnsfilter.Config{
		SafeBrowsingEnabled:   true,
		SafeBrowsingCacheSize: 1000,
		SafeSearchEnabled:     true,
		SafeSearchCacheSize:   1000,
		ParentalCacheSize:     1000,
		CacheTime:             30,
	}
	forwardConf := ServerConfig{
		UDPListenAddrs: []*net.UDPAddr{},
		TCPListenAddrs: []*net.TCPAddr{},
		FilteringConfig: FilteringConfig{
			ProtectionEnabled: true,
			UpstreamDNS:       []string{"8.8.8.8:53", "8.8.4.4:53"},
		},
		ConfigModified: func() {},
	}
	s := createTestServer(t, filterConf, forwardConf)

	defaultConf := s.conf

	err := s.Start()
	assert.Nil(t, err)
	t.Cleanup(func() {
		assert.Nil(t, s.Stop())
	})

	w := httptest.NewRecorder()

	const casesFileName = "handleSetConfig_cases.json"
	var testCases []struct {
		Name    string `json:"name"`
		Req     string `json:"req"`
		WantSet string `json:"wantSet"`
		WantGet string `json:"wantGet"`
	}
	openCases(t, casesFileName, &testCases)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Cleanup(func() {
				s.conf = defaultConf
			})

			rBody := ioutil.NopCloser(strings.NewReader(tc.Req))
			var r *http.Request
			r, err = http.NewRequest(http.MethodPost, "http://example.com", rBody)
			require.Nil(t, err)

			s.handleSetConfig(w, r)
			assert.Equal(t, tc.WantSet, w.Body.String())
			w.Body.Reset()

			s.handleGetConfig(w, nil)
			assert.Equal(t, tc.WantGet, w.Body.String())
			w.Body.Reset()
		})
	}
}

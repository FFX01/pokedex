package pokecache

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
    var tests = []time.Duration{
        (time.Second * 10),
        (time.Second * 20),
        (time.Second * 1),
    }

    for _, tt := range tests {
        testname := fmt.Sprintf("TestNewCache duration %v", tt)
        t.Run(testname, func(t *testing.T) {
            c := NewCache(tt)
            if c.interval != tt {
                t.Errorf("Expected interval to be %v but got %v", tt, c.interval)
            }
        })
    }
}

func TestAddAndGet(t *testing.T) {
    var tests = []struct{
        url string
        data []byte
    }{
        {"https://example.com", []byte("some data")},
        {"https://example.com?page=2", []byte("page 2 data")},
        {"https://example.com?page=3&limit=100", []byte("page 3 data with limit")},
    }

    cache := NewCache(time.Second * 60)

    for _, tt := range tests {
        cache.Add(tt.url, tt.data)
        data, found := cache.Get(tt.url)
        if !found {
            t.Errorf("Expected %s to be in cache, but it was not found", tt.url)
        }
        if !reflect.DeepEqual(data, tt.data) {
            t.Errorf("Expected data %s but found %s", tt.data, data)
        }
    }
}

func TestTTL(t *testing.T) {
    type TTLTest struct {
        url string
        data []byte
        ttl int
    }

    var tests = []TTLTest{
        {"https://example.com/one-second", []byte("one second"), 1},
        {"https://example.come/two-seconds", []byte("two seconds"), 2},
        {"https://example.com/three-seconds", []byte("three seconds"), 3},
    }

    for _, tt := range tests {
        go func() {
            cache := NewCache(time.Second * time.Duration(tt.ttl))
            cache.Add(tt.url, tt.data)
            time.Sleep(time.Duration(tt.ttl + 1) * time.Second)
            result, found := cache.Get(tt.url)
            if found {
                t.Errorf("Expected cache key %s to be empty but found data", tt.url)
            }
            if !reflect.DeepEqual(result, []byte{}) {
                t.Errorf("Expected empty byte slice but got %v", result)
            }
        }()
    }
}

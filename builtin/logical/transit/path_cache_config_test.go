package transit

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

const targetCacheSize = 12345

func TestTransit_CacheConfig(t *testing.T) {
	b1, storage1 := createBackendWithSysView(t)

	doReq := func(b *backend, req *logical.Request) *logical.Response {
		resp, err := b.HandleRequest(context.Background(), req)
		if err != nil || (resp != nil && resp.IsError()) {
			t.Fatalf("got err:\n%#v\nreq:\n%#v\n", err, *req)
		}
		return resp
	}
	// doErrReq := func(req *logical.Request) {
	// 	resp, err := b.HandleRequest(context.Background(), req)
	// 	if err == nil {
	// 		if resp == nil || !resp.IsError() {
	// 			t.Fatalf("expected error; req:\n%#v\n", *req)
	// 		}
	// 	}
	// }

	validateResponse := func(resp *logical.Response, expectedCacheSize int, expectedWarning bool) {
		actualCacheSize, ok := resp.Data["cache_size"].(int)
		if !ok {
			t.Fatalf("No cache_size returned")
		}
		if expectedCacheSize != actualCacheSize {
			t.Fatalf("testAccReadCacheConfig expected: %d got: %d", expectedCacheSize, actualCacheSize)
		}
		// check for the presence/absence of warnings - warnings are expected if a cache size has been configured but
		// not yet applied by reloading the mount
		warningCheckPass := expectedWarning == (len(resp.Warnings) > 0)
		if !warningCheckPass {
			t.Fatalf(
				"testAccSteporeadCacheConfig warnings error.\nexpect warnings: %t but number of warnings was: %d",
				expectedWarning, len(resp.Warnings),
			)
		}
	}

	writeReq := &logical.Request{
		Storage:   storage1,
		Operation: logical.UpdateOperation,
		Path:      "cache-config",
		Data: map[string]interface{}{
			"size": targetCacheSize,
		},
	}

	readReq := &logical.Request{
		Storage:   storage1,
		Operation: logical.ReadOperation,
		Path:      "cache-config",
	}

	// test stuff
	// default cache should be zero
	validateResponse(doReq(b1, readReq), 0, false)
	doReq(b1, writeReq)
	validateResponse(doReq(b1, readReq), targetCacheSize, true)

	b2, _ := createBackendWithSysView(t)
	fmt.Printf("%t", b1 == b2)
	validateResponse(doReq(b2, readReq), targetCacheSize, false)

	// fmt.Printf("resp: %#v", doReq(readReq))
}

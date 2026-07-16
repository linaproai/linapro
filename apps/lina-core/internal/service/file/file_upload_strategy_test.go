// This file unit-tests automatic upload strategy planning.

package file

import "testing"

func TestPlanUploadStrategyBelowThresholdPrefersDirectSingle(t *testing.T) {
	t.Parallel()
	got := planUploadStrategy(10*bytesPerMegabyte, true, 100, true, true)
	if got.Channel != UploadChannelDirect || got.Encoding != UploadEncodingSingle {
		t.Fatalf("got %+v", got)
	}
}

func TestPlanUploadStrategyAboveThresholdDirectMultipart(t *testing.T) {
	t.Parallel()
	got := planUploadStrategy(150*bytesPerMegabyte, true, 100, true, true)
	if got.Channel != UploadChannelDirect || got.Encoding != UploadEncodingMultipart {
		t.Fatalf("got %+v", got)
	}
}

func TestPlanUploadStrategyAboveThresholdProxyChunkedWithoutCloudMP(t *testing.T) {
	t.Parallel()
	got := planUploadStrategy(150*bytesPerMegabyte, true, 100, false, false)
	if got.Channel != UploadChannelProxy || got.Encoding != UploadEncodingMultipart {
		t.Fatalf("got %+v", got)
	}
	// Direct put available but no cloud multipart still falls back to proxy chunked.
	got = planUploadStrategy(150*bytesPerMegabyte, true, 100, true, false)
	if got.Channel != UploadChannelProxy || got.Encoding != UploadEncodingMultipart {
		t.Fatalf("got %+v", got)
	}
}

func TestPlanUploadStrategyDisabledMultipart(t *testing.T) {
	t.Parallel()
	got := planUploadStrategy(150*bytesPerMegabyte, false, 100, true, true)
	if got.Channel != UploadChannelDirect || got.Encoding != UploadEncodingSingle {
		t.Fatalf("got %+v", got)
	}
}

func TestBuildMultipartPlanEnforcesMinPartSize(t *testing.T) {
	t.Parallel()
	plan := buildMultipartPlan(3, 0)
	if plan.PartSize != 5*bytesPerMegabyte {
		t.Fatalf("part size: got %d", plan.PartSize)
	}
	if plan.MaxConcurrency != 1 {
		t.Fatalf("concurrency: got %d", plan.MaxConcurrency)
	}
}

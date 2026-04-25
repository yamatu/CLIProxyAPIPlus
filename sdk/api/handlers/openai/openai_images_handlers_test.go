package openai

import (
	"context"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/interfaces"
	"github.com/tidwall/gjson"
)

func TestBuildImagesResponsesRequestUsesGPTImage2Tool(t *testing.T) {
	tool := []byte(`{"type":"image_generation","action":"generate","model":"gpt-image-2","size":"1024x1024"}`)
	got := buildImagesResponsesRequest("draw a cat", nil, tool)

	if model := gjson.GetBytes(got, "model").String(); model != defaultImagesMainModel {
		t.Fatalf("model = %q, want %q", model, defaultImagesMainModel)
	}
	if toolModel := gjson.GetBytes(got, "tools.0.model").String(); toolModel != defaultImagesToolModel {
		t.Fatalf("tools.0.model = %q, want %q", toolModel, defaultImagesToolModel)
	}
	if toolType := gjson.GetBytes(got, "tools.0.type").String(); toolType != "image_generation" {
		t.Fatalf("tools.0.type = %q, want image_generation", toolType)
	}
	if prompt := gjson.GetBytes(got, "input.0.content.0.text").String(); prompt != "draw a cat" {
		t.Fatalf("prompt = %q, want draw a cat", prompt)
	}
}

func TestCollectImagesFromResponsesStreamBuildsImagesAPIResponse(t *testing.T) {
	data := make(chan []byte, 1)
	errs := make(chan *interfaces.ErrorMessage)
	data <- []byte(`data: {"type":"response.completed","response":{"created_at":1776902400,"output":[{"type":"image_generation_call","result":"abc123","output_format":"png","size":"1024x1024"}],"tool_usage":{"image_gen":{"input_tokens":1,"output_tokens":2,"total_tokens":3}}}}`)
	close(data)
	close(errs)

	got, errMsg := collectImagesFromResponsesStream(context.Background(), data, errs, "b64_json")
	if errMsg != nil {
		t.Fatalf("collectImagesFromResponsesStream returned error: %v", errMsg.Error)
	}
	if b64 := gjson.GetBytes(got, "data.0.b64_json").String(); b64 != "abc123" {
		t.Fatalf("b64_json = %q, want abc123", b64)
	}
	if outputFormat := gjson.GetBytes(got, "output_format").String(); outputFormat != "png" {
		t.Fatalf("output_format = %q, want png", outputFormat)
	}
	if total := gjson.GetBytes(got, "usage.total_tokens").Int(); total != 3 {
		t.Fatalf("usage.total_tokens = %d, want 3", total)
	}
}

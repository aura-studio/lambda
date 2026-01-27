package tests

import (
	"bytes"
	"testing"

	"github.com/aura-studio/lambda/invoke"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// **Feature: invoke-lambda-handler, Property 1: 请求响应往返一致性**
// **Validates: Requirements 4.8**
//
// Property 1: 请求响应往返一致性
// For any 有效的 Request 对象，序列化后再反序列化 SHALL 产生等价的对象。
// 同样，For any 有效的 Response 对象，序列化后再反序列化 SHALL 产生等价的对象。

// genRequest generates random valid Request objects for property testing
func genRequest() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString(),                    // correlation_id
		gen.AlphaString().Map(func(s string) string { return "/" + s }), // path (ensure it starts with /)
		gen.SliceOf(gen.UInt8()),             // payload
	).Map(func(values []interface{}) *invoke.Request {
		return &invoke.Request{
			CorrelationId: values[0].(string),
			Path:          values[1].(string),
			Payload:       values[2].([]byte),
		}
	})
}

// genResponse generates random valid Response objects for property testing
func genResponse() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString(),        // correlation_id
		gen.SliceOf(gen.UInt8()), // payload
		gen.AlphaString(),        // error
	).Map(func(values []interface{}) *invoke.Response {
		return &invoke.Response{
			CorrelationId: values[0].(string),
			Payload:       values[1].([]byte),
			Error:         values[2].(string),
		}
	})
}

// requestsEqual checks if two Request objects are equivalent
func requestsEqual(a, b *invoke.Request) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.CorrelationId == b.CorrelationId &&
		a.Path == b.Path &&
		bytes.Equal(a.Payload, b.Payload)
}

// responsesEqual checks if two Response objects are equivalent
func responsesEqual(a, b *invoke.Response) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.CorrelationId == b.CorrelationId &&
		bytes.Equal(a.Payload, b.Payload) &&
		a.Error == b.Error
}

// TestRequestRoundTrip tests that Request objects can be serialized and deserialized
// without losing data (round-trip consistency).
// **Validates: Requirements 4.8**
func TestRequestRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Request round-trip: serialize then deserialize produces equivalent object", prop.ForAll(
		func(original *invoke.Request) bool {
			// Serialize the request
			data, err := proto.Marshal(original)
			if err != nil {
				t.Logf("Failed to marshal Request: %v", err)
				return false
			}

			// Deserialize the request
			deserialized := &invoke.Request{}
			err = proto.Unmarshal(data, deserialized)
			if err != nil {
				t.Logf("Failed to unmarshal Request: %v", err)
				return false
			}

			// Verify equivalence
			if !requestsEqual(original, deserialized) {
				t.Logf("Request mismatch:\nOriginal: %+v\nDeserialized: %+v", original, deserialized)
				return false
			}

			return true
		},
		genRequest(),
	))

	properties.TestingRun(t)
}

// TestResponseRoundTrip tests that Response objects can be serialized and deserialized
// without losing data (round-trip consistency).
// **Validates: Requirements 4.8**
func TestResponseRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Response round-trip: serialize then deserialize produces equivalent object", prop.ForAll(
		func(original *invoke.Response) bool {
			// Serialize the response
			data, err := proto.Marshal(original)
			if err != nil {
				t.Logf("Failed to marshal Response: %v", err)
				return false
			}

			// Deserialize the response
			deserialized := &invoke.Response{}
			err = proto.Unmarshal(data, deserialized)
			if err != nil {
				t.Logf("Failed to unmarshal Response: %v", err)
				return false
			}

			// Verify equivalence
			if !responsesEqual(original, deserialized) {
				t.Logf("Response mismatch:\nOriginal: %+v\nDeserialized: %+v", original, deserialized)
				return false
			}

			return true
		},
		genResponse(),
	))

	properties.TestingRun(t)
}

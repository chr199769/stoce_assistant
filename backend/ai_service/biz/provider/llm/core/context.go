package core

import "context"

type modelNameKey struct{}
type generationNameKey struct{}
type generationMetadataKey struct{}

func WithModelName(ctx context.Context, modelName string) context.Context {
	if modelName == "" {
		return ctx
	}
	return context.WithValue(ctx, modelNameKey{}, modelName)
}

func ModelNameFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val := ctx.Value(modelNameKey{})
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func WithGenerationName(ctx context.Context, name string) context.Context {
	if name == "" {
		return ctx
	}
	return context.WithValue(ctx, generationNameKey{}, name)
}

func GenerationNameFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val := ctx.Value(generationNameKey{})
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func WithGenerationMetadata(ctx context.Context, metadata map[string]interface{}) context.Context {
	if metadata == nil || len(metadata) == 0 {
		return ctx
	}
	return context.WithValue(ctx, generationMetadataKey{}, metadata)
}

func GenerationMetadataFromContext(ctx context.Context) map[string]interface{} {
	if ctx == nil {
		return nil
	}
	val := ctx.Value(generationMetadataKey{})
	if val == nil {
		return nil
	}
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	return nil
}

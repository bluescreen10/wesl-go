package ast

// IsBuiltinType reports whether name is a WGSL built-in type identifier.
// It returns true for scalar types, vector and matrix types, texture types,
// sampler types, and other standard predeclared types defined by the WGSL
// specification.
func IsBuiltinType(name string) bool {
	switch name {
	case "array", "atomic", "bitcast", "bool",
		"f16", "f32", "i32", "u32",
		"mat2x2", "mat2x3", "mat2x4",
		"mat3x2", "mat3x3", "mat3x4",
		"mat4x2", "mat4x3", "mat4x4",
		"ptr",
		"sampler", "sampler_comparison",
		"texture_1d", "texture_2d", "texture_2d_array",
		"texture_3d", "texture_cube", "texture_cube_array",
		"texture_depth_2d", "texture_depth_2d_array",
		"texture_depth_cube", "texture_depth_cube_array",
		"texture_depth_multisampled_2d",
		"texture_multisampled_2d",
		"texture_storage_1d", "texture_storage_2d",
		"texture_storage_2d_array", "texture_storage_3d",
		"vec2", "vec3", "vec4",
		"binding_array":
		return true
	default:
		return false
	}
}

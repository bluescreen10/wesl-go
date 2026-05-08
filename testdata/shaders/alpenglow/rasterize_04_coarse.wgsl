@group(0) @binding(0)
var<uniform> config: TwoPassConfig;
@group(0) @binding(1)
var<storage, read> coarse_renderable_faces: array<TwoPassCoarseRenderableFace>;
@group(0) @binding(2)
var<storage, read> coarse_edges: array<LinearEdge>;
@group(0) @binding(3)
var<storage, read_write> fine_renderable_faces: array<TwoPassFineRenderableFace>;
@group(0) @binding(4)
var<storage, read_write> fine_edges: array<LinearEdge>;
@group(0) @binding(5)
var<storage, read_write> addresses: array<atomic<u32>>;
@group(0) @binding(6)
var<storage, read> tile_addresses: array<u32>;
struct TwoPassCoarseRenderableFace {
  bits: u32,
  edges_index: u32,
  num_edges: u32,
  clip_counts: u32,
  tile_index: u32
}
struct LinearEdge {
  startPoint: vec2f,
  endPoint: vec2f
}
struct bounds_clip_edge_Result {
  edges: array<LinearEdge,3>,
  count: u32
}
struct MD_ClipResult {
  p0: vec2f,
  p1: vec2f,
  clipped: bool
}
fn matthes_drakopoulos_clip(
  p0: vec2f,
  p1: vec2f,
  minX: f32,
  minY: f32,
  maxX: f32,
  maxY: f32
) -> MD_ClipResult {
  if (
    !( p0.x < minX && p1.x < minX ) &&
    !( p0.x > maxX && p1.x > maxX ) &&
    !( p0.y < minY && p1.y < minY ) &&
    !( p0.y > maxY && p1.y > maxY )
  ) {
    var x0 = p0.x;
    var y0 = p0.y;
    var x1 = p1.x;
    var y1 = p1.y;
    if ( x0 < minX ) {
      x0 = minX;
      y0 = ( p1.y - p0.y ) * ( minX - p0.x ) / ( p1.x - p0.x ) + p0.y;
    }
    else if ( x0 > maxX ) {
      x0 = maxX;
      y0 = ( p1.y - p0.y ) * ( maxX - p0.x ) / ( p1.x - p0.x ) + p0.y;
    }
    if ( y0 < minY ) {
      y0 = minY;
      x0 = ( p1.x - p0.x ) * ( minY - p0.y ) / ( p1.y - p0.y ) + p0.x;
    }
    else if ( y0 > maxY ) {
      y0 = maxY;
      x0 = ( p1.x - p0.x ) * ( maxY - p0.y ) / ( p1.y - p0.y ) + p0.x;
    }
    if ( x1 < minX ) {
      x1 = minX;
      y1 = ( p1.y - p0.y ) * ( minX - p0.x ) / ( p1.x - p0.x ) + p0.y;
    }
    else if ( x1 > maxX ) {
      x1 = maxX;
      y1 = ( p1.y - p0.y ) * ( maxX - p0.x ) / ( p1.x - p0.x ) + p0.y;
    }
    if ( y1 < minY ) {
      y1 = minY;
      x1 = ( p1.x - p0.x ) * ( minY - p0.y ) / ( p1.y - p0.y ) + p0.x;
    }
    else if ( y1 > maxY ) {
      y1 = maxY;
      x1 = ( p1.x - p0.x ) * ( maxY - p0.y ) / ( p1.y - p0.y ) + p0.x;
    }
    if ( !( x0 < minX && x1 < minX ) && !( x0 > maxX && x1 > maxX ) ) {
      return MD_ClipResult( vec2( x0, y0 ), vec2( x1, y1 ), true );
    }
  }
  return MD_ClipResult( p0, p1, false );
}
fn bounds_clip_edge(
  edge: LinearEdge,
  minX: f32,
  minY: f32,
  maxX: f32,
  maxY: f32,
  centerX: f32,
  centerY: f32
) -> bounds_clip_edge_Result {
  var edges: array<LinearEdge,3>;
  var count: u32 = 0u;
  let startPoint = edge.startPoint;
  let endPoint = edge.endPoint;
  let clipResult: MD_ClipResult = matthes_drakopoulos_clip( startPoint, endPoint, minX, minY, maxX, maxY );
  let clippedStartPoint = clipResult.p0;
  let clippedEndPoint = clipResult.p1;
  let clipped = clipResult.clipped;
  let startXLess = startPoint.x < centerX;
  let startYLess = startPoint.y < centerY;
  let endXLess = endPoint.x < centerX;
  let endYLess = endPoint.y < centerY;
  let needsStartCorner = !clipped || !all( startPoint == clippedStartPoint );
  let needsEndCorner = !clipped || !all( endPoint == clippedEndPoint );
  let startCorner = vec2(
    select( maxX, minX, startXLess ),
    select( maxY, minY, startYLess )
  );
  let endCorner = vec2(
    select( maxX, minX, endXLess ),
    select( maxY, minY, endYLess )
  );
  if ( clipped ) {
    if ( needsStartCorner && !all( startCorner == clippedStartPoint ) ) {
      edges[ count ] = LinearEdge( startCorner, clippedStartPoint );
      count++;
    }
    if ( !all( clippedStartPoint == clippedEndPoint ) ) {
      edges[ count ] = LinearEdge( clippedStartPoint, clippedEndPoint );
      count++;
    }
    if ( needsEndCorner && !all( endCorner == clippedEndPoint ) ) {
      edges[ count ] = LinearEdge( clippedEndPoint, endCorner );
      count++;
    }
  }
  else {
    if ( startXLess != endXLess && startYLess != endYLess ) {
      let y = startPoint.y + ( endPoint.y - startPoint.y ) * ( centerX - startPoint.x ) / ( endPoint.x - startPoint.x );
      let startSame = startXLess == startYLess;
      let yGreater = y > centerY;
      let middlePoint = vec2(
        select( maxX, minX, startSame == yGreater ),
        select( minY, maxY, yGreater )
      );
      edges[ 0u ] = LinearEdge( startCorner, middlePoint );
      edges[ 1u ] = LinearEdge( middlePoint, endCorner );
      count = 2u;
    }
    else if ( !all( startCorner == endCorner ) ) {
      edges[ 0u ] = LinearEdge( startCorner, endCorner );
      count = 1u;
    }
  }
  return bounds_clip_edge_Result( edges, count );
}
struct TwoPassFineRenderableFace {
  bits: u32,
  edges_index: u32,
  num_edges: u32,
  clip_counts: u32,
  next_address: u32
}
const low_area_multiplier = 0.0002f;
var<workgroup> coarse_face: TwoPassCoarseRenderableFace;
var<workgroup> scratch_data: array<vec2u, 256>;
var<workgroup> base_indices: vec2u;
@compute @workgroup_size(256)
fn main(
  @builtin(global_invocation_id) global_id: vec3u,
  @builtin(local_invocation_id) local_id: vec3u,
  @builtin(workgroup_id) workgroup_id: vec3u
) {
  if ( workgroup_id.x >= tile_addresses[ 0u ] ) {
    return;
  }
  if ( local_id.x == 0u ) {
    coarse_face = coarse_renderable_faces[ workgroup_id.x ];
  }
  workgroupBarrier();
  let tile_index_xy = vec2(
    coarse_face.tile_index % config.tile_width,
    coarse_face.tile_index / config.tile_width
  );
  let tile_xy = tile_index_xy * vec2( config.tile_size );
  let relative_bin_xy = vec2( local_id.x % 16u, local_id.x / 16u );
  let filter_expansion = select( select( 2f, 1f, config.filter_type == 1u ), 0.5f, config.filter_type == 0u ) * config.filter_scale - 0.5f;
  let min = vec2f( tile_xy + vec2( config.bin_size ) * relative_bin_xy ) - vec2( filter_expansion );
  let max = vec2f( tile_xy + vec2( config.bin_size ) * ( vec2( 1u ) + relative_bin_xy ) ) + vec2( filter_expansion );
  var area: f32;
  var num_clipped_edges: u32 = 0u;
  var clipped_clip_counts = unpack4xI8( coarse_face.clip_counts );
  let max_area = ( max.x - min.x ) * ( max.y - min.y );
  let is_source_full_area = ( coarse_face.bits & 0x80000000u ) != 0u;
  if ( is_source_full_area ) {
    area = max_area;
  }
  else {
    let source_clip_counts = vec4f( clipped_clip_counts );
    area = 2f * ( max.y - min.y ) * ( source_clip_counts.x * min.x + source_clip_counts.z * max.x );
    for ( var edge_offset = 0u; edge_offset < coarse_face.num_edges; edge_offset++ ) {
      let linear_edge = coarse_edges[ coarse_face.edges_index + edge_offset ];
      let bounds_centroid = 0.5f * ( min + max );
      let result = bounds_clip_edge( linear_edge, min.x, min.y, max.x, max.y, bounds_centroid.x, bounds_centroid.y );
      for ( var i = 0u; i < result.count; i++ ) {
        let edge = result.edges[ i ];
        let p0 = edge.startPoint;
        let p1 = edge.endPoint;
        area += ( p1.x + p0.x ) * ( p1.y - p0.y );
        if ( is_edge_clipped_count( p0, p1, min, max ) ) {
          let count_delta = select( -1i, 1i, p0.x + p0.y < p1.x + p1.y );
          let index = select( select( 3u, 1u, p0.y == min.y ), select( 2u, 0u, p0.x == min.x ), p0.x == p1.x );
          clipped_clip_counts[ index ] += count_delta;
        }
        else {
          num_clipped_edges += 1u;
        }
      }
    }
    area *= 0.5f;
  }
  let is_full_area = is_source_full_area || area + low_area_multiplier >= max_area;
  let needs_write_face = area > low_area_multiplier && ( num_clipped_edges > 0u || clipped_clip_counts[ 0u ] != 0i || clipped_clip_counts[ 1u ] != 0i || clipped_clip_counts[ 2u ] != 0i || clipped_clip_counts[ 3u ] != 0i );
  let needs_write_edges = needs_write_face && !is_full_area;
  let required_edge_count = select( 0u, num_clipped_edges, needs_write_edges );
  let required_face_count = select( 0u, 1u, needs_write_face );
  var offsets = vec2( required_edge_count, required_face_count );
  /*** begin scan direction:left exclusive:false ***/
  /*** loading scratch ***/
  scratch_data[ local_id.x ] = offsets;
  workgroupBarrier();
  if ( local_id.x >= 1u ) {
    offsets = ( scratch_data[ local_id.x - 1u ] + offsets );
  }
  workgroupBarrier();
  scratch_data[ local_id.x ] = offsets;
  workgroupBarrier();
  if ( local_id.x >= 2u ) {
    offsets = ( scratch_data[ local_id.x - 2u ] + offsets );
  }
  workgroupBarrier();
  scratch_data[ local_id.x ] = offsets;
  workgroupBarrier();
  if ( local_id.x >= 4u ) {
    offsets = ( scratch_data[ local_id.x - 4u ] + offsets );
  }
  workgroupBarrier();
  scratch_data[ local_id.x ] = offsets;
  workgroupBarrier();
  if ( local_id.x >= 8u ) {
    offsets = ( scratch_data[ local_id.x - 8u ] + offsets );
  }
  workgroupBarrier();
  scratch_data[ local_id.x ] = offsets;
  workgroupBarrier();
  if ( local_id.x >= 16u ) {
    offsets = ( scratch_data[ local_id.x - 16u ] + offsets );
  }
  workgroupBarrier();
  scratch_data[ local_id.x ] = offsets;
  workgroupBarrier();
  if ( local_id.x >= 32u ) {
    offsets = ( scratch_data[ local_id.x - 32u ] + offsets );
  }
  workgroupBarrier();
  scratch_data[ local_id.x ] = offsets;
  workgroupBarrier();
  if ( local_id.x >= 64u ) {
    offsets = ( scratch_data[ local_id.x - 64u ] + offsets );
  }
  workgroupBarrier();
  scratch_data[ local_id.x ] = offsets;
  workgroupBarrier();
  if ( local_id.x >= 128u ) {
    offsets = ( scratch_data[ local_id.x - 128u ] + offsets );
  }
  /*** end scan ***/
  if ( local_id.x == 0xffu ) {
    base_indices = vec2(
      atomicAdd( &addresses[ 1u ], offsets.x ),
      atomicAdd( &addresses[ 0u ], offsets.y )
    );
  }
  workgroupBarrier();
  let edges_index = base_indices.x + offsets.x - required_edge_count;
  let face_index = base_indices.y + offsets.y - required_face_count;
  if ( !needs_write_face ) {
    return;
  }
  let bin_index = local_id.x + ( coarse_face.tile_index << 8u );
  let previous_address = atomicExchange( &addresses[ bin_index + 2u ], face_index );
  fine_renderable_faces[ face_index ] = TwoPassFineRenderableFace(
    coarse_face.bits | select( 0u, 0x80000000u, is_full_area ),
    edges_index,
    required_edge_count,
    pack4xI8( clipped_clip_counts ),
    previous_address
  );
  if ( !needs_write_edges ) {
    return;
  }
  var edge_index = edges_index;
  for ( var edge_offset = 0u; edge_offset < coarse_face.num_edges; edge_offset++ ) {
    let linear_edge = coarse_edges[ coarse_face.edges_index + edge_offset ];
    let bounds_centroid = 0.5f * ( min + max );
    let result = bounds_clip_edge( linear_edge, min.x, min.y, max.x, max.y, bounds_centroid.x, bounds_centroid.y );
    for ( var i = 0u; i < result.count; i++ ) {
      let edge = result.edges[ i ];
      if ( !is_edge_clipped_count( edge.startPoint, edge.endPoint, min, max ) ) {
        fine_edges[ edge_index ] = edge;
        edge_index += 1u;
      }
    }
  }
}
fn is_edge_clipped_count( p0: vec2f, p1: vec2f, min: vec2f, max: vec2f ) -> bool {
  return all( ( p0 == min ) | ( p0 == max ) ) &&
  all( ( p1 == min ) | ( p1 == max ) ) &&
  ( p0.x == p1.x ) != ( p0.y == p1.y );
}
struct TwoPassConfig {
  raster_width: u32,
  raster_height: u32,
  tile_width: u32,
  tile_height: u32,
  bin_width: u32,
  bin_height: u32,
  tile_size: u32,
  bin_size: u32,
  filter_type: u32,
  filter_scale: f32,
  raster_color_space: u32
}

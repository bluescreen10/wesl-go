@group(0) @binding(0)
var<storage, read_write> data: array<vec2u, 3>;
var<workgroup> scratch: array<vec2u, 64>;
@compute @workgroup_size(32)
fn main(
  @builtin(global_invocation_id) global_id: vec3u,
  @builtin(local_invocation_id) local_id: vec3u,
  @builtin(workgroup_id) workgroup_id: vec3u
) {
  /*** begin scan_comprehensive inclusive ***/
  /*** begin load_multiple ***/
  {
    let base_striped_index = workgroup_id.x * 64u + local_id.x;
    {
      let striped_index = base_striped_index + 0u;
      var lm_val: vec2u;
      if ( striped_index < ( ( ( 6209u ) + 4095u ) / 4096u ) ) {
        lm_val = data[ striped_index ];
      }
      else {
        lm_val = vec2( 0u );
      }
      scratch[ striped_index - workgroup_id.x * 64u ] = lm_val;
    }
    {
      let striped_index = base_striped_index + 32u;
      var lm_val: vec2u;
      if ( striped_index < ( ( ( 6209u ) + 4095u ) / 4096u ) ) {
        lm_val = data[ striped_index ];
      }
      else {
        lm_val = vec2( 0u );
      }
      scratch[ striped_index - workgroup_id.x * 64u ] = lm_val;
    }
  }
  /*** end load_multiple ***/
  workgroupBarrier();
  /*** begin scan_raked ***/
  /*** begin (sequential scan of tile) ***/
  var value = scratch[ local_id.x * 2u ];
  {
    value = ( value + scratch[ local_id.x * 2u + 1u ] - min( value.y, scratch[ local_id.x * 2u + 1u ].x ) );
    scratch[ local_id.x * 2u + 1u ] = value;
  }
  /*** end (sequential scan of tile) ***/
  workgroupBarrier();
  /*** begin scan direction:left exclusive:false ***/
  if ( local_id.x >= 1u ) {
    value = ( scratch[ ( local_id.x - 1u ) * 2u + 1u ] + value - min( scratch[ ( local_id.x - 1u ) * 2u + 1u ].y, value.x ) );
  }
  workgroupBarrier();
  scratch[ ( local_id.x ) * 2u + 1u ] = value;
  workgroupBarrier();
  if ( local_id.x >= 2u ) {
    value = ( scratch[ ( local_id.x - 2u ) * 2u + 1u ] + value - min( scratch[ ( local_id.x - 2u ) * 2u + 1u ].y, value.x ) );
  }
  workgroupBarrier();
  scratch[ ( local_id.x ) * 2u + 1u ] = value;
  workgroupBarrier();
  if ( local_id.x >= 4u ) {
    value = ( scratch[ ( local_id.x - 4u ) * 2u + 1u ] + value - min( scratch[ ( local_id.x - 4u ) * 2u + 1u ].y, value.x ) );
  }
  workgroupBarrier();
  scratch[ ( local_id.x ) * 2u + 1u ] = value;
  workgroupBarrier();
  if ( local_id.x >= 8u ) {
    value = ( scratch[ ( local_id.x - 8u ) * 2u + 1u ] + value - min( scratch[ ( local_id.x - 8u ) * 2u + 1u ].y, value.x ) );
  }
  workgroupBarrier();
  scratch[ ( local_id.x ) * 2u + 1u ] = value;
  workgroupBarrier();
  if ( local_id.x >= 16u ) {
    value = ( scratch[ ( local_id.x - 16u ) * 2u + 1u ] + value - min( scratch[ ( local_id.x - 16u ) * 2u + 1u ].y, value.x ) );
  }
  workgroupBarrier();
  scratch[ ( local_id.x ) * 2u + 1u ] = value;
  /*** end scan ***/
  workgroupBarrier();
  /*** begin (add scanned values to tile) ***/
  var added_value = select( vec2( 0u ), scratch[ local_id.x * 2u - 1u ], local_id.x > 0 );
  {
    let index = local_id.x * 2u + 0u;
    var current_value: vec2u;
    current_value = ( added_value + scratch[ index ] - min( added_value.y, scratch[ index ].x ) );
    scratch[ index ] = current_value;
  }
  /*** end (add scanned values to tile) ***/
  /*** end scan_raked ***/
  workgroupBarrier();
  /*** begin (output write) ***/
  {
    let coalesced_local_index = 0u + local_id.x;
    let coalesced_data_index = workgroup_id.x * 64u + coalesced_local_index;
    if ( coalesced_data_index < ( ( ( 6209u ) + 4095u ) / 4096u ) ) {
      data[ coalesced_data_index ] = scratch[ coalesced_local_index ];
    }
  }
  {
    let coalesced_local_index = 32u + local_id.x;
    let coalesced_data_index = workgroup_id.x * 64u + coalesced_local_index;
    if ( coalesced_data_index < ( ( ( 6209u ) + 4095u ) / 4096u ) ) {
      data[ coalesced_data_index ] = scratch[ coalesced_local_index ];
    }
  }
  /*** end (output write) ***/
  /*** end scan_comprehensive ***/
}

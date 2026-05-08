@group(0) @binding(0)
var<storage, read> input: array<vec2u, 131072>;
@group(0) @binding(1)
var<storage, read> histogram_offsets: array<u32, 131072>;
@group(0) @binding(2)
var<storage, read_write> output: array<vec2u, 131072>;
var<workgroup> bits_scratch: array<u32, 64>;
var<workgroup> value_scratch: array<vec2u, 256>;
var<workgroup> local_histogram_offsets: array<u32, 256u>;
var<workgroup> start_indices: array<u32, 256>;
@compute @workgroup_size(64)
fn main(
  @builtin(global_invocation_id) global_id: vec3u,
  @builtin(local_invocation_id) local_id: vec3u,
  @builtin(workgroup_id) workgroup_id: vec3u
) {
  let num_valid_workgroups = ( ( ( 131045u ) + 255u ) / 256u );
  if ( workgroup_id.x < num_valid_workgroups ) {
    /*** begin load_multiple ***/
    {
      let base_striped_index = workgroup_id.x * 256u + local_id.x;
      {
        let striped_index = base_striped_index + 0u;
        if ( striped_index < 131045u ) {
          var lm_val: vec2u;
          lm_val = input[ striped_index ];
          value_scratch[ striped_index - workgroup_id.x * 256u ] = lm_val;
        }
      }
      {
        let striped_index = base_striped_index + 64u;
        if ( striped_index < 131045u ) {
          var lm_val: vec2u;
          lm_val = input[ striped_index ];
          value_scratch[ striped_index - workgroup_id.x * 256u ] = lm_val;
        }
      }
      {
        let striped_index = base_striped_index + 128u;
        if ( striped_index < 131045u ) {
          var lm_val: vec2u;
          lm_val = input[ striped_index ];
          value_scratch[ striped_index - workgroup_id.x * 256u ] = lm_val;
        }
      }
      {
        let striped_index = base_striped_index + 192u;
        if ( striped_index < 131045u ) {
          var lm_val: vec2u;
          lm_val = input[ striped_index ];
          value_scratch[ striped_index - workgroup_id.x * 256u ] = lm_val;
        }
      }
    }
    /*** end load_multiple ***/
    /*** begin load histogram offsets ***/
    {
      let local_index = 0u + local_id.x;
      if ( local_index < 256u ) {
        local_histogram_offsets[ local_index ] = histogram_offsets[ local_index * num_valid_workgroups + workgroup_id.x ];
      }
    }
    {
      let local_index = 64u + local_id.x;
      if ( local_index < 256u ) {
        local_histogram_offsets[ local_index ] = histogram_offsets[ local_index * num_valid_workgroups + workgroup_id.x ];
      }
    }
    {
      let local_index = 128u + local_id.x;
      if ( local_index < 256u ) {
        local_histogram_offsets[ local_index ] = histogram_offsets[ local_index * num_valid_workgroups + workgroup_id.x ];
      }
    }
    {
      let local_index = 192u + local_id.x;
      if ( local_index < 256u ) {
        local_histogram_offsets[ local_index ] = histogram_offsets[ local_index * num_valid_workgroups + workgroup_id.x ];
      }
    }
    /*** end load histogram offsets ***/
    workgroupBarrier();
    let reduced_length = ( 131045u ) - workgroup_id.x * 256u;
    for ( var srs_i = 0u; srs_i < 8u; srs_i += 2u ) {
      /*** begin n_bit_compact_single_sort ***/
      {
        var tb_bits_vector = 0u;
        if ( 4u * local_id.x + 0u < reduced_length ) {
          let tb_value = value_scratch[ 4u * local_id.x + 0u ];
          let tb_bits = ( ( ( ( ( tb_value.y >> 0u ) & 255u ) ) >> srs_i ) & 3u );
          tb_bits_vector += 1u << ( ( ( tb_bits ) % 4u ) * 8u );
        }
        if ( 4u * local_id.x + 1u < reduced_length ) {
          let tb_value = value_scratch[ 4u * local_id.x + 1u ];
          let tb_bits = ( ( ( ( ( tb_value.y >> 0u ) & 255u ) ) >> srs_i ) & 3u );
          tb_bits_vector += 1u << ( ( ( tb_bits ) % 4u ) * 8u );
        }
        if ( 4u * local_id.x + 2u < reduced_length ) {
          let tb_value = value_scratch[ 4u * local_id.x + 2u ];
          let tb_bits = ( ( ( ( ( tb_value.y >> 0u ) & 255u ) ) >> srs_i ) & 3u );
          tb_bits_vector += 1u << ( ( ( tb_bits ) % 4u ) * 8u );
        }
        if ( 4u * local_id.x + 3u < reduced_length ) {
          let tb_value = value_scratch[ 4u * local_id.x + 3u ];
          let tb_bits = ( ( ( ( ( tb_value.y >> 0u ) & 255u ) ) >> srs_i ) & 3u );
          tb_bits_vector += 1u << ( ( ( tb_bits ) % 4u ) * 8u );
        }
        /*** begin scan direction:left exclusive:true ***/
        /*** loading scratch ***/
        bits_scratch[ local_id.x ] = tb_bits_vector;
        workgroupBarrier();
        if ( local_id.x >= 1u ) {
          tb_bits_vector = ( bits_scratch[ local_id.x - 1u ] + tb_bits_vector );
        }
        workgroupBarrier();
        bits_scratch[ local_id.x ] = tb_bits_vector;
        workgroupBarrier();
        if ( local_id.x >= 2u ) {
          tb_bits_vector = ( bits_scratch[ local_id.x - 2u ] + tb_bits_vector );
        }
        workgroupBarrier();
        bits_scratch[ local_id.x ] = tb_bits_vector;
        workgroupBarrier();
        if ( local_id.x >= 4u ) {
          tb_bits_vector = ( bits_scratch[ local_id.x - 4u ] + tb_bits_vector );
        }
        workgroupBarrier();
        bits_scratch[ local_id.x ] = tb_bits_vector;
        workgroupBarrier();
        if ( local_id.x >= 8u ) {
          tb_bits_vector = ( bits_scratch[ local_id.x - 8u ] + tb_bits_vector );
        }
        workgroupBarrier();
        bits_scratch[ local_id.x ] = tb_bits_vector;
        workgroupBarrier();
        if ( local_id.x >= 16u ) {
          tb_bits_vector = ( bits_scratch[ local_id.x - 16u ] + tb_bits_vector );
        }
        workgroupBarrier();
        bits_scratch[ local_id.x ] = tb_bits_vector;
        workgroupBarrier();
        if ( local_id.x >= 32u ) {
          tb_bits_vector = ( bits_scratch[ local_id.x - 32u ] + tb_bits_vector );
        }
        workgroupBarrier();
        bits_scratch[ local_id.x ] = tb_bits_vector;
        workgroupBarrier();
        tb_bits_vector = select( 0u, bits_scratch[ local_id.x - 1u ], local_id.x > 0u );
        /*** end scan ***/
        var tb_offsets = bits_scratch[ 63u ];
        /*** begin bit_pack_radix_exclusive_scan ***/
        var bitty_value = 0u;
        var bitty_next_value = 0u;
        bitty_next_value += tb_offsets & 0xffu;
        tb_offsets = ( tb_offsets & 0xffffff00u ) | bitty_value;
        bitty_value = bitty_next_value;
        bitty_next_value += ( tb_offsets >> 8u ) & 0xffu;
        tb_offsets = ( tb_offsets & 0xffff00ffu ) | ( ( bitty_value ) << 8u );
        bitty_value = bitty_next_value;
        bitty_next_value += ( tb_offsets >> 16u ) & 0xffu;
        tb_offsets = ( tb_offsets & 0xff00ffffu ) | ( ( bitty_value ) << 16u );
        bitty_value = bitty_next_value;
        tb_offsets = ( tb_offsets & 0xffffffu ) | ( ( bitty_value ) << 24u );
        /*** end bit_pack_radix_exclusive_scan ***/
        var tb_values: array<vec2u, 4>;
        tb_values[ 0u ] = value_scratch[ 4u * local_id.x + 0u ];
        tb_values[ 1u ] = value_scratch[ 4u * local_id.x + 1u ];
        tb_values[ 2u ] = value_scratch[ 4u * local_id.x + 2u ];
        tb_values[ 3u ] = value_scratch[ 4u * local_id.x + 3u ];
        workgroupBarrier();
        if ( 4u * local_id.x + 0u < reduced_length ) {
          let tb_value = tb_values[ 0u ];
          let tb_bits = ( ( ( ( ( tb_value.y >> 0u ) & 255u ) ) >> srs_i ) & 3u );
          value_scratch[ ( ( tb_offsets >> ( ( ( tb_bits ) % 4u ) * 8u ) ) & 255u ) + ( ( tb_bits_vector >> ( ( ( tb_bits ) % 4u ) * 8u ) ) & 255u ) ] = tb_value;
          tb_bits_vector += 1u << ( ( ( tb_bits ) % 4u ) * 8u );
        }
        if ( 4u * local_id.x + 1u < reduced_length ) {
          let tb_value = tb_values[ 1u ];
          let tb_bits = ( ( ( ( ( tb_value.y >> 0u ) & 255u ) ) >> srs_i ) & 3u );
          value_scratch[ ( ( tb_offsets >> ( ( ( tb_bits ) % 4u ) * 8u ) ) & 255u ) + ( ( tb_bits_vector >> ( ( ( tb_bits ) % 4u ) * 8u ) ) & 255u ) ] = tb_value;
          tb_bits_vector += 1u << ( ( ( tb_bits ) % 4u ) * 8u );
        }
        if ( 4u * local_id.x + 2u < reduced_length ) {
          let tb_value = tb_values[ 2u ];
          let tb_bits = ( ( ( ( ( tb_value.y >> 0u ) & 255u ) ) >> srs_i ) & 3u );
          value_scratch[ ( ( tb_offsets >> ( ( ( tb_bits ) % 4u ) * 8u ) ) & 255u ) + ( ( tb_bits_vector >> ( ( ( tb_bits ) % 4u ) * 8u ) ) & 255u ) ] = tb_value;
          tb_bits_vector += 1u << ( ( ( tb_bits ) % 4u ) * 8u );
        }
        if ( 4u * local_id.x + 3u < reduced_length ) {
          let tb_value = tb_values[ 3u ];
          let tb_bits = ( ( ( ( ( tb_value.y >> 0u ) & 255u ) ) >> srs_i ) & 3u );
          value_scratch[ ( ( tb_offsets >> ( ( ( tb_bits ) % 4u ) * 8u ) ) & 255u ) + ( ( tb_bits_vector >> ( ( ( tb_bits ) % 4u ) * 8u ) ) & 255u ) ] = tb_value;
          tb_bits_vector += 1u << ( ( ( tb_bits ) % 4u ) * 8u );
        }
        workgroupBarrier();
      }
      /*** end n_bit_compact_single_sort ***/
    }
    /*** begin write start_indices ***/
    {
      let local_index = 0u + local_id.x;
      if ( local_index < reduced_length ) {
        var head_value = 0u;
        if ( local_index > 0u && ( ( value_scratch[ local_index ].y >> 0u ) & 255u ) != ( ( value_scratch[ local_index - 1u ].y >> 0u ) & 255u ) ) {
          head_value = local_index;
        }
        start_indices[ local_index ] = head_value;
      }
    }
    {
      let local_index = 64u + local_id.x;
      if ( local_index < reduced_length ) {
        var head_value = 0u;
        if ( local_index > 0u && ( ( value_scratch[ local_index ].y >> 0u ) & 255u ) != ( ( value_scratch[ local_index - 1u ].y >> 0u ) & 255u ) ) {
          head_value = local_index;
        }
        start_indices[ local_index ] = head_value;
      }
    }
    {
      let local_index = 128u + local_id.x;
      if ( local_index < reduced_length ) {
        var head_value = 0u;
        if ( local_index > 0u && ( ( value_scratch[ local_index ].y >> 0u ) & 255u ) != ( ( value_scratch[ local_index - 1u ].y >> 0u ) & 255u ) ) {
          head_value = local_index;
        }
        start_indices[ local_index ] = head_value;
      }
    }
    {
      let local_index = 192u + local_id.x;
      if ( local_index < reduced_length ) {
        var head_value = 0u;
        if ( local_index > 0u && ( ( value_scratch[ local_index ].y >> 0u ) & 255u ) != ( ( value_scratch[ local_index - 1u ].y >> 0u ) & 255u ) ) {
          head_value = local_index;
        }
        start_indices[ local_index ] = head_value;
      }
    }
    /*** end write start_indices ***/
    workgroupBarrier();
    /*** begin scan_raked ***/
    /*** begin (sequential scan of tile) ***/
    var value = start_indices[ local_id.x * 4u ];
    {
      value = max( value, start_indices[ local_id.x * 4u + 1u ] );
      start_indices[ local_id.x * 4u + 1u ] = value;
    }
    {
      value = max( value, start_indices[ local_id.x * 4u + 2u ] );
      start_indices[ local_id.x * 4u + 2u ] = value;
    }
    {
      value = max( value, start_indices[ local_id.x * 4u + 3u ] );
      start_indices[ local_id.x * 4u + 3u ] = value;
    }
    /*** end (sequential scan of tile) ***/
    workgroupBarrier();
    /*** begin scan direction:left exclusive:false ***/
    if ( local_id.x >= 1u ) {
      value = max( start_indices[ ( local_id.x - 1u ) * 4u + 3u ], value );
    }
    workgroupBarrier();
    start_indices[ ( local_id.x ) * 4u + 3u ] = value;
    workgroupBarrier();
    if ( local_id.x >= 2u ) {
      value = max( start_indices[ ( local_id.x - 2u ) * 4u + 3u ], value );
    }
    workgroupBarrier();
    start_indices[ ( local_id.x ) * 4u + 3u ] = value;
    workgroupBarrier();
    if ( local_id.x >= 4u ) {
      value = max( start_indices[ ( local_id.x - 4u ) * 4u + 3u ], value );
    }
    workgroupBarrier();
    start_indices[ ( local_id.x ) * 4u + 3u ] = value;
    workgroupBarrier();
    if ( local_id.x >= 8u ) {
      value = max( start_indices[ ( local_id.x - 8u ) * 4u + 3u ], value );
    }
    workgroupBarrier();
    start_indices[ ( local_id.x ) * 4u + 3u ] = value;
    workgroupBarrier();
    if ( local_id.x >= 16u ) {
      value = max( start_indices[ ( local_id.x - 16u ) * 4u + 3u ], value );
    }
    workgroupBarrier();
    start_indices[ ( local_id.x ) * 4u + 3u ] = value;
    workgroupBarrier();
    if ( local_id.x >= 32u ) {
      value = max( start_indices[ ( local_id.x - 32u ) * 4u + 3u ], value );
    }
    workgroupBarrier();
    start_indices[ ( local_id.x ) * 4u + 3u ] = value;
    /*** end scan ***/
    workgroupBarrier();
    /*** begin (add scanned values to tile) ***/
    var added_value = select( 0u, start_indices[ local_id.x * 4u - 1u ], local_id.x > 0 );
    {
      let index = local_id.x * 4u + 0u;
      var current_value: u32;
      current_value = max( added_value, start_indices[ index ] );
      start_indices[ index ] = current_value;
    }
    {
      let index = local_id.x * 4u + 1u;
      var current_value: u32;
      current_value = max( added_value, start_indices[ index ] );
      start_indices[ index ] = current_value;
    }
    {
      let index = local_id.x * 4u + 2u;
      var current_value: u32;
      current_value = max( added_value, start_indices[ index ] );
      start_indices[ index ] = current_value;
    }
    /*** end (add scanned values to tile) ***/
    /*** end scan_raked ***/
    workgroupBarrier();
    /*** begin write output ***/
    {
      let local_index = 0u + local_id.x;
      if ( local_index < reduced_length ) {
        let local_offset = local_index - start_indices[ local_index ];
        let value = value_scratch[ local_index ];
        let offset = local_histogram_offsets[ ( ( value.y >> 0u ) & 255u ) ] + local_offset;
        output[ offset ] = value;
      }
    }
    {
      let local_index = 64u + local_id.x;
      if ( local_index < reduced_length ) {
        let local_offset = local_index - start_indices[ local_index ];
        let value = value_scratch[ local_index ];
        let offset = local_histogram_offsets[ ( ( value.y >> 0u ) & 255u ) ] + local_offset;
        output[ offset ] = value;
      }
    }
    {
      let local_index = 128u + local_id.x;
      if ( local_index < reduced_length ) {
        let local_offset = local_index - start_indices[ local_index ];
        let value = value_scratch[ local_index ];
        let offset = local_histogram_offsets[ ( ( value.y >> 0u ) & 255u ) ] + local_offset;
        output[ offset ] = value;
      }
    }
    {
      let local_index = 192u + local_id.x;
      if ( local_index < reduced_length ) {
        let local_offset = local_index - start_indices[ local_index ];
        let value = value_scratch[ local_index ];
        let offset = local_histogram_offsets[ ( ( value.y >> 0u ) & 255u ) ] + local_offset;
        output[ offset ] = value;
      }
    }
    /*** end write output ***/
  }
}

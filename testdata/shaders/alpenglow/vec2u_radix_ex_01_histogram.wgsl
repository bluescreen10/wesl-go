@group(0) @binding(0)
var<storage, read> input: array<vec2u, 131072>;
@group(0) @binding(1)
var<storage, read_write> output: array<u32, 131072>;
var<workgroup> histogram_scratch: array<atomic<u32>, 256>;
@compute @workgroup_size(64)
fn main(
  @builtin(global_invocation_id) global_id: vec3u,
  @builtin(local_invocation_id) local_id: vec3u,
  @builtin(workgroup_id) workgroup_id: vec3u
) {
  /*** begin radix_histogram ***/
  {
    /*** begin histogram ***/
    {
      {
        let coalesced_local_index = 0u + local_id.x;
        let coalesced_data_index = workgroup_id.x * 256u + coalesced_local_index;
        if ( coalesced_data_index < 131045u ) {
          atomicAdd( &histogram_scratch[ ( ( input[ coalesced_data_index ].y >> 0u ) & 255u ) ], 1u );
        }
      }
      {
        let coalesced_local_index = 64u + local_id.x;
        let coalesced_data_index = workgroup_id.x * 256u + coalesced_local_index;
        if ( coalesced_data_index < 131045u ) {
          atomicAdd( &histogram_scratch[ ( ( input[ coalesced_data_index ].y >> 0u ) & 255u ) ], 1u );
        }
      }
      {
        let coalesced_local_index = 128u + local_id.x;
        let coalesced_data_index = workgroup_id.x * 256u + coalesced_local_index;
        if ( coalesced_data_index < 131045u ) {
          atomicAdd( &histogram_scratch[ ( ( input[ coalesced_data_index ].y >> 0u ) & 255u ) ], 1u );
        }
      }
      {
        let coalesced_local_index = 192u + local_id.x;
        let coalesced_data_index = workgroup_id.x * 256u + coalesced_local_index;
        if ( coalesced_data_index < 131045u ) {
          atomicAdd( &histogram_scratch[ ( ( input[ coalesced_data_index ].y >> 0u ) & 255u ) ], 1u );
        }
      }
    }
    /*** end histogram ***/
    let num_valid_workgroups = ( ( ( 131045u ) + 255u ) / 256u );
    if ( workgroup_id.x < num_valid_workgroups ) {
      workgroupBarrier();
      {
        let local_index = 0u + local_id.x;
        if ( local_index < 256u ) {
          output[ local_index * num_valid_workgroups + workgroup_id.x ] = atomicLoad( &histogram_scratch[ local_index ] );
        }
      }
      {
        let local_index = 64u + local_id.x;
        if ( local_index < 256u ) {
          output[ local_index * num_valid_workgroups + workgroup_id.x ] = atomicLoad( &histogram_scratch[ local_index ] );
        }
      }
      {
        let local_index = 128u + local_id.x;
        if ( local_index < 256u ) {
          output[ local_index * num_valid_workgroups + workgroup_id.x ] = atomicLoad( &histogram_scratch[ local_index ] );
        }
      }
      {
        let local_index = 192u + local_id.x;
        if ( local_index < 256u ) {
          output[ local_index * num_valid_workgroups + workgroup_id.x ] = atomicLoad( &histogram_scratch[ local_index ] );
        }
      }
    }
  }
  /*** end radix_histogram ***/
}

@group(0) @binding(0)
var<storage, read> mhist_input: array<u32, 2000>;
@group(0) @binding(1)
var<storage, read_write> mhist_output: array<atomic<u32>, 256>;
var<workgroup> histogram_scratch: array<atomic<u32>, 256>;
@compute @workgroup_size(64)
fn main(
  @builtin(global_invocation_id) global_id: vec3u,
  @builtin(local_invocation_id) local_id: vec3u,
  @builtin(workgroup_id) workgroup_id: vec3u
) {
  /*** begin histogram ***/
  {
    {
      let coalesced_local_index = 0u + local_id.x;
      let coalesced_data_index = workgroup_id.x * 512u + coalesced_local_index;
      if ( coalesced_data_index < 2000u ) {
        atomicAdd( &histogram_scratch[ ( mhist_input[ coalesced_data_index ] % 256u ) ], 1u );
      }
    }
    {
      let coalesced_local_index = 64u + local_id.x;
      let coalesced_data_index = workgroup_id.x * 512u + coalesced_local_index;
      if ( coalesced_data_index < 2000u ) {
        atomicAdd( &histogram_scratch[ ( mhist_input[ coalesced_data_index ] % 256u ) ], 1u );
      }
    }
    {
      let coalesced_local_index = 128u + local_id.x;
      let coalesced_data_index = workgroup_id.x * 512u + coalesced_local_index;
      if ( coalesced_data_index < 2000u ) {
        atomicAdd( &histogram_scratch[ ( mhist_input[ coalesced_data_index ] % 256u ) ], 1u );
      }
    }
    {
      let coalesced_local_index = 192u + local_id.x;
      let coalesced_data_index = workgroup_id.x * 512u + coalesced_local_index;
      if ( coalesced_data_index < 2000u ) {
        atomicAdd( &histogram_scratch[ ( mhist_input[ coalesced_data_index ] % 256u ) ], 1u );
      }
    }
    {
      let coalesced_local_index = 256u + local_id.x;
      let coalesced_data_index = workgroup_id.x * 512u + coalesced_local_index;
      if ( coalesced_data_index < 2000u ) {
        atomicAdd( &histogram_scratch[ ( mhist_input[ coalesced_data_index ] % 256u ) ], 1u );
      }
    }
    {
      let coalesced_local_index = 320u + local_id.x;
      let coalesced_data_index = workgroup_id.x * 512u + coalesced_local_index;
      if ( coalesced_data_index < 2000u ) {
        atomicAdd( &histogram_scratch[ ( mhist_input[ coalesced_data_index ] % 256u ) ], 1u );
      }
    }
    {
      let coalesced_local_index = 384u + local_id.x;
      let coalesced_data_index = workgroup_id.x * 512u + coalesced_local_index;
      if ( coalesced_data_index < 2000u ) {
        atomicAdd( &histogram_scratch[ ( mhist_input[ coalesced_data_index ] % 256u ) ], 1u );
      }
    }
    {
      let coalesced_local_index = 448u + local_id.x;
      let coalesced_data_index = workgroup_id.x * 512u + coalesced_local_index;
      if ( coalesced_data_index < 2000u ) {
        atomicAdd( &histogram_scratch[ ( mhist_input[ coalesced_data_index ] % 256u ) ], 1u );
      }
    }
  }
  /*** end histogram ***/
  workgroupBarrier();
  {
    let mhist_index = 0u + local_id.x;
    if ( mhist_index < 256u ) {
      atomicAdd( &mhist_output[ mhist_index ], atomicLoad( &histogram_scratch[ mhist_index ] ) );
    }
  }
  {
    let mhist_index = 64u + local_id.x;
    if ( mhist_index < 256u ) {
      atomicAdd( &mhist_output[ mhist_index ], atomicLoad( &histogram_scratch[ mhist_index ] ) );
    }
  }
  {
    let mhist_index = 128u + local_id.x;
    if ( mhist_index < 256u ) {
      atomicAdd( &mhist_output[ mhist_index ], atomicLoad( &histogram_scratch[ mhist_index ] ) );
    }
  }
  {
    let mhist_index = 192u + local_id.x;
    if ( mhist_index < 256u ) {
      atomicAdd( &mhist_output[ mhist_index ], atomicLoad( &histogram_scratch[ mhist_index ] ) );
    }
  }
}

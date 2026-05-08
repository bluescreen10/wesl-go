@group(0) @binding(0)
var<storage, read> input: array<i32, 56>;
@group(0) @binding(1)
var<storage, read_write> output: array<u32, 224>;
alias q128 = vec4<u32>;
struct IntersectionPoint {
  t0: q128,
  t1: q128,
  px: q128,
  py: q128
}
struct LineSegmentIntersection {
  num_points: u32,
  p0: IntersectionPoint,
  p1: IntersectionPoint
}
alias i64 = vec2<u32>;
fn i32_to_i64( x: i32 ) -> i64 {
  return vec2<u32>( u32( x ), select( 0u, 0xffffffffu, x < 0i ) );
}
alias u64 = vec2<u32>;
fn add_u32_u32_to_u64( a: u32, b: u32 ) -> u64 {
  let sum = a + b;
  return vec2( sum, select( 0u, 1u, sum < a ) );
}
fn add_u64_u64( a: u64, b: u64 ) -> u64 {
  return add_u32_u32_to_u64( a.x, b.x ) + vec2( 0u, add_u32_u32_to_u64( a.y, b.y ).x );
}
const ONE_u64 = vec2( 1u, 0u );
fn mul_u32_u32_to_u64( a: u32, b: u32 ) -> u64 {
  let a_low = a & 0xffffu;
  let a_high = a >> 16u;
  let b_low = b & 0xffffu;
  let b_high = b >> 16u;
  let c_low = a_low * b_low;
  let c_mid = a_low * b_high + a_high * b_low;
  let c_high = a_high * b_high;
  let low = add_u32_u32_to_u64( c_low, c_mid << 16u );
  let high = vec2( 0u, ( c_mid >> 16u ) + c_high );
  return low + high;
}
fn mul_u64_u64( a: u64, b: u64 ) -> u64 {
  let low = mul_u32_u32_to_u64( a.x, b.x );
  let mid0 = vec2( 0u, mul_u32_u32_to_u64( a.x, b.y ).x );
  let mid1 = vec2( 0u, mul_u32_u32_to_u64( a.y, b.x ).x );
  return add_u64_u64( add_u64_u64( low, mid0 ), mid1 );
}
fn abs_i64( a: i64 ) -> i64 {
  return select( a, add_u64_u64( ~( a ), ONE_u64 ), ( ( ( a ).y >> 31u ) == 1u ) );
}
fn mul_i64_i64( a: i64, b: i64 ) -> i64 {
  var result = mul_u64_u64( abs_i64( a ), abs_i64( b ) );
  result.y &= 0x7fffffffu;
  if ( ( ( ( a ).y >> 31u ) == 1u ) != ( ( ( b ).y >> 31u ) == 1u ) ) {
    return add_u64_u64( ~( result ), ONE_u64 );
  }
  else {
    return result;
  }
}
fn is_zero_u64( a: u64 ) -> bool {
  return a.x == 0u && a.y == 0u;
}
fn i64_to_q128( numerator: i64, denominator: i64 ) -> q128 {
  if ( ( ( ( denominator ).y >> 31u ) == 1u ) ) {
    return vec4( add_u64_u64( ~( numerator ), ONE_u64 ), add_u64_u64( ~( denominator ), ONE_u64 ) );
  }
  else {
    return vec4( numerator, denominator );
  }
}
fn equals_cross_mul_q128( a: q128, b: q128 ) -> bool {
  return is_zero_u64( add_u64_u64( mul_i64_i64( a.xy, b.zw ), add_u64_u64( ~( mul_i64_i64( a.zw, b.xy ) ), ONE_u64 ) ) );
}
fn is_zero_q128( a: q128 ) -> bool {
  return a.x == 0u && a.y == 0u;
}
fn cmp_i64_i64( a: i64, b: i64 ) -> i32 {
  let diff = add_u64_u64( a, add_u64_u64( ~( b ), ONE_u64 ) );
  if ( is_zero_u64( diff ) ) {
    return 0i;
  }
  else {
    return select( 1i, -1i, ( ( ( diff ).y >> 31u ) == 1u ) );
  }
}
const ZERO_u64 = vec2( 0u, 0u );
fn ratio_test_q128( q: q128 ) -> i32 {
  return cmp_i64_i64( q.xy, ZERO_u64 ) + cmp_i64_i64( q.zw, q.xy );
}
const ZERO_q128 = vec4( 0u, 0u, 1u, 0u );
fn first_trailing_bit_u64( a: u64 ) -> u32 {
  if ( a.x != 0u ) {
    return firstTrailingBit( a.x );
  }
  else {
    return firstTrailingBit( a.y ) + 32u;
  }
}
fn right_shift_u64( a: u64, b: u32 ) -> u64 {
  if ( b == 0u ) {
    return a;
  }
  else if ( b < 32u ) {
    return vec2( ( a.x >> b ) | ( a.y << ( 32u - b ) ), a.y >> b );
  }
  else {
    return vec2( a.y >> ( b - 32u ), 0u );
  }
}
fn cmp_u64_u64( a: u64, b: u64 ) -> i32 {
  if ( a.y < b.y ) {
    return -1i;
  }
  else if ( a.y > b.y ) {
    return 1i;
  }
  else {
    if ( a.x < b.x ) {
      return -1i;
    }
    else if ( a.x > b.x ) {
      return 1i;
    }
    else {
      return 0i;
    }
  }
}
fn left_shift_u64( a: u64, b: u32 ) -> u64 {
  if ( b == 0u ) {
    return a;
  }
  else if ( b < 32u ) {
    return vec2( a.x << b, ( a.y << b ) | ( a.x >> ( 32u - b ) ) );
  }
  else {
    return vec2( 0u, a.x << ( b - 32u ) );
  }
}
fn gcd_u64_u64( a: u64, b: u64 ) -> u64 {
  if ( is_zero_u64( a ) ) {
    return b;
  }
  else if ( is_zero_u64( b ) ) {
    return a;
  }
  let gcd_two = first_trailing_bit_u64( a | b );
  var u = right_shift_u64( a, gcd_two );
  var v = right_shift_u64( b, gcd_two );
  while ( u.x != v.x || u.y != v.y ) {
    if ( cmp_u64_u64( u, v ) == -1i ) {
      let t = u;
      u = v;
      v = t;
    }
    u = add_u64_u64( u, add_u64_u64( ~( v ), ONE_u64 ) );
    u = right_shift_u64( u, first_trailing_bit_u64( u ) );
  }
  return left_shift_u64( u, gcd_two );
}
fn first_leading_bit_u64( a: u64 ) -> u32 {
  if ( a.y != 0u ) {
    return firstLeadingBit( a.y ) + 32u;
  }
  else {
    return firstLeadingBit( a.x );
  }
}
fn div_u64_u64( a: u64, b: u64 ) -> vec4<u32> {
  if ( is_zero_u64( a ) ) {
    return vec4( 0u, 0u, 0u, 0u );
  }
  else if ( is_zero_u64( b ) ) {
    return vec4( 0u, 0u, 0u, 0u );
  }
  var result = vec2( 0u, 0u );
  var remainder = a;
  let high_bit = min( first_leading_bit_u64( a ), first_leading_bit_u64( b ) );
  var count = 63u - high_bit;
  var divisor = left_shift_u64( b, count );
  while( !is_zero_u64( remainder ) ) {
    if ( cmp_u64_u64( remainder, divisor ) >= 0i ) {
      remainder = add_u64_u64( remainder, add_u64_u64( ~( divisor ), ONE_u64 ) );
      result = result | left_shift_u64( ONE_u64, count );
    }
    if ( count == 0u ) {
      break;
    }
    divisor = right_shift_u64( divisor, 1u );
    count -= 1u;
  }
  return vec4( result, remainder );
}
fn reduce_q128( a: q128 ) -> q128 {
  let numerator = a.xy;
  let denominator = a.zw;
  if ( numerator.x == 0u && numerator.y == 0u ) {
    return vec4( 0u, 0u, 1u, 0u );
  }
  else if ( denominator.x == 1 && denominator.y == 0u ) {
    return a;
  }
  let abs_numerator = abs_i64( numerator );
  let gcd = gcd_u64_u64( abs_numerator, denominator );
  if ( gcd.x == 1u && gcd.y == 0u ) {
    return a;
  }
  else {
    let reduced_numerator = div_u64_u64( abs_numerator, gcd ).xy;
    let reduced_denominator = div_u64_u64( denominator, gcd ).xy;
    if ( ( ( ( numerator ).y >> 31u ) == 1u ) ) {
      return vec4( add_u64_u64( ~( reduced_numerator ), ONE_u64 ), reduced_denominator );
    }
    else {
      return vec4( reduced_numerator, reduced_denominator );
    }
  }
}
const ONE_q128 = vec4( 1u, 0u, 1u, 0u );
const not_rational = vec4( 0u, 0u, 0u, 0u );
const not_point = IntersectionPoint( not_rational, not_rational, not_rational, not_rational );
const not_intersection = LineSegmentIntersection( 0u, not_point, not_point );
fn intersect_line_segments( p0: vec2i, p1: vec2i, p2: vec2i, p3: vec2i ) -> LineSegmentIntersection {
  let p0x = i32_to_i64( p0.x );
  let p0y = i32_to_i64( p0.y );
  let p1x = i32_to_i64( p1.x );
  let p1y = i32_to_i64( p1.y );
  let p2x = i32_to_i64( p2.x );
  let p2y = i32_to_i64( p2.y );
  let p3x = i32_to_i64( p3.x );
  let p3y = i32_to_i64( p3.y );
  let d0x = add_u64_u64( p1x, add_u64_u64( ~( p0x ), ONE_u64 ) );
  let d0y = add_u64_u64( p1y, add_u64_u64( ~( p0y ), ONE_u64 ) );
  let d1x = add_u64_u64( p3x, add_u64_u64( ~( p2x ), ONE_u64 ) );
  let d1y = add_u64_u64( p3y, add_u64_u64( ~( p2y ), ONE_u64 ) );
  let cdx = add_u64_u64( p2x, add_u64_u64( ~( p0x ), ONE_u64 ) );
  let cdy = add_u64_u64( p2y, add_u64_u64( ~( p0y ), ONE_u64 ) );
  let denominator = add_u64_u64( mul_i64_i64( d0x, d1y ), add_u64_u64( ~( mul_i64_i64( d0y, d1x ) ), ONE_u64 ) );
  if ( is_zero_u64( denominator ) ) {
    var a: q128;
    var b: q128;
    let d1x_zero = is_zero_u64( d1x );
    let d1y_zero = is_zero_u64( d1y );
    if ( d1x_zero && d1y_zero ) {
      return not_intersection;
    }
    else if ( d1x_zero ) {
      if ( p0.x != p2.x ) {
        return not_intersection;
      }
      a = i64_to_q128( d0y, d1y );
      b = i64_to_q128( add_u64_u64( ~( cdy ), ONE_u64 ), d1y );
    }
    else if ( d1y_zero ) {
      if ( p0.y != p2.y ) {
        return not_intersection;
      }
      a = i64_to_q128( d0x, d1x );
      b = i64_to_q128( add_u64_u64( ~( cdx ), ONE_u64 ), d1x );
    }
    else {
      if ( is_zero_u64( d0x ) && is_zero_u64( d0y ) ) {
        return not_intersection;
      }
      let ax = i64_to_q128( d0x, d1x );
      let ay = i64_to_q128( d0y, d1y );
      if ( !equals_cross_mul_q128( ax, ay ) ) {
        return not_intersection;
      }
      let bx = i64_to_q128( add_u64_u64( ~( cdx ), ONE_u64 ), d1x );
      let by = i64_to_q128( add_u64_u64( ~( cdy ), ONE_u64 ), d1y );
      if ( !equals_cross_mul_q128( bx, by ) ) {
        return not_intersection;
      }
      if ( is_zero_q128( ax ) ) {
        a = ay;
        b = by;
      }
      else {
        a = ax;
        b = bx;
      }
    }
    var points: u32 = 0u;
    var results = array<IntersectionPoint, 2u>( not_point, not_point );
    let case1t1 = b;
    if ( ratio_test_q128( case1t1 ) == 2i ) {
      let p = IntersectionPoint( ZERO_q128, reduce_q128( case1t1 ), vec4( p0x, 1u, 0u ), vec4( p0y, 1u, 0u ) );
      results[ points ] = p;
      points += 1u;
    }
    let case2t1 = vec4( add_u64_u64( a.xy, b.xy ), a.zw );
    if ( ratio_test_q128( case2t1 ) == 2i ) {
      let p = IntersectionPoint( ONE_q128, reduce_q128( case2t1 ), vec4( p1x, 1u, 0u ), vec4( p1y, 1u, 0u ) );
      results[ points ] = p;
      points += 1u;
    }
    let case3t0 = i64_to_q128( add_u64_u64( ~( b.xy ), ONE_u64 ), a.xy );
    if ( ratio_test_q128( case3t0 ) == 2i ) {
      let p = IntersectionPoint( reduce_q128( case3t0 ), ZERO_q128, vec4( p2x, 1u, 0u ), vec4( p2y, 1u, 0u ) );
      results[ points ] = p;
      points += 1u;
    }
    let case4t0 = i64_to_q128( add_u64_u64( a.zw, add_u64_u64( ~( b.xy ), ONE_u64 ) ), a.xy );
    if ( ratio_test_q128( case4t0 ) == 2i ) {
      let p = IntersectionPoint( reduce_q128( case4t0 ), ONE_q128, vec4( p3x, 1u, 0u ), vec4( p3y, 1u, 0u ) );
      results[ points ] = p;
      points += 1u;
    }
    return LineSegmentIntersection( points, results[ 0 ], results[ 1 ] );
  }
  else {
    let t_numerator = add_u64_u64( mul_i64_i64( cdx, d1y ), add_u64_u64( ~( mul_i64_i64( cdy, d1x ) ), ONE_u64 ) );
    let u_numerator = add_u64_u64( mul_i64_i64( cdx, d0y ), add_u64_u64( ~( mul_i64_i64( cdy, d0x ) ), ONE_u64 ) );
    let t_raw = i64_to_q128( t_numerator, denominator );
    let u_raw = i64_to_q128( u_numerator, denominator );
    let t_cmp = ratio_test_q128( t_raw );
    let u_cmp = ratio_test_q128( u_raw );
    if ( t_cmp <= 0i || u_cmp <= 0i ) {
      return not_intersection;
    }
    else if ( t_cmp == 1i && u_cmp == 1i ) {
      return not_intersection;
    }
    else {
      let x_numerator = add_u64_u64( mul_i64_i64( denominator, p0x ), mul_i64_i64( t_numerator, d0x ) );
      let y_numerator = add_u64_u64( mul_i64_i64( denominator, p0y ), mul_i64_i64( t_numerator, d0y ) );
      let x_raw = i64_to_q128( x_numerator, denominator );
      let y_raw = i64_to_q128( y_numerator, denominator );
      let x = reduce_q128( x_raw );
      let y = reduce_q128( y_raw );
      let t = reduce_q128( t_raw );
      let u = reduce_q128( u_raw );
      return LineSegmentIntersection( 1u, IntersectionPoint( t, u, x, y ), not_point );
    }
  }
}
@compute @workgroup_size(1) fn main(
  @builtin(global_invocation_id) id: vec3<u32>
) {
  let i = id.x;
  let in = i * 8u;
  let out = i * 32u;
  let p0 = vec2( input[ in + 0u ], input[ in + 1u ] );
  let p1 = vec2( input[ in + 2u ], input[ in + 3u ] );
  let p2 = vec2( input[ in + 4u ], input[ in + 5u ] );
  let p3 = vec2( input[ in + 6u ], input[ in + 7u ] );
  let c = intersect_line_segments( p0, p1, p2, p3 );
  output[ out + 0u ] = c.p0.t0.x;
  output[ out + 1u ] = c.p0.t0.y;
  output[ out + 2u ] = c.p0.t0.z;
  output[ out + 3u ] = c.p0.t0.w;
  output[ out + 4u ] = c.p0.t1.x;
  output[ out + 5u ] = c.p0.t1.y;
  output[ out + 6u ] = c.p0.t1.z;
  output[ out + 7u ] = c.p0.t1.w;
  output[ out + 8u ] = c.p0.px.x;
  output[ out + 9u ] = c.p0.px.y;
  output[ out + 10u ] = c.p0.px.z;
  output[ out + 11u ] = c.p0.px.w;
  output[ out + 12u ] = c.p0.py.x;
  output[ out + 13u ] = c.p0.py.y;
  output[ out + 14u ] = c.p0.py.z;
  output[ out + 15u ] = c.p0.py.w;
  output[ out + 16u ] = c.p1.t0.x;
  output[ out + 17u ] = c.p1.t0.y;
  output[ out + 18u ] = c.p1.t0.z;
  output[ out + 19u ] = c.p1.t0.w;
  output[ out + 20u ] = c.p1.t1.x;
  output[ out + 21u ] = c.p1.t1.y;
  output[ out + 22u ] = c.p1.t1.z;
  output[ out + 23u ] = c.p1.t1.w;
  output[ out + 24u ] = c.p1.px.x;
  output[ out + 25u ] = c.p1.px.y;
  output[ out + 26u ] = c.p1.px.z;
  output[ out + 27u ] = c.p1.px.w;
  output[ out + 28u ] = c.p1.py.x;
  output[ out + 29u ] = c.p1.py.y;
  output[ out + 30u ] = c.p1.py.z;
  output[ out + 31u ] = c.p1.py.w;
}

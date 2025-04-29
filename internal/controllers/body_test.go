package controllers_test

var bodyCreateShortURLJSONListSUCCESS string = `[
  {
    "correlation_id": "a1b2c3d4e5f6g7h8",
    "original_url": "https://astrodata.io/reports/yearly-forecast"
  },
  {
    "correlation_id": "z9y8x7w6v5u4t3s2",
    "original_url": "https://bluemelon.net/gallery/9f3d1"
  },
  {
    "correlation_id": "m1n2b3v4c5x6z7l8",
    "original_url": "https://cloudyfox.org/blogs/hidden-pathways"
  },
  {
    "correlation_id": "q1w2e3r4t5y6u7i8",
    "original_url": "https://lunarbyte.com/downloads/setup_v2.1.exe"
  },
  {
    "correlation_id": "p0o9i8u7y6t5r4e3",
    "original_url": "https://pinetreehub.com/shop/category/garden-tools"
  },
  {
    "correlation_id": "k9j8h7g6f5d4s3a2",
    "original_url": "https://sparkledger.app/api/v1/transactions"
  },
  {
    "correlation_id": "u7y6t5r4e3w2q1z0",
    "original_url": "https://mochakitty.co/forum/topic/1873"
  },
  {
    "correlation_id": "x1c2v3b4n5m6l7k8",
    "original_url": "https://velvetwave.net/music/playlist/lofi-sunset"
  },
  {
    "correlation_id": "s9d8f7g6h5j4k3l2",
    "original_url": "https://oceanbrewery.org/events/summerfest2025"
  },
  {
    "correlation_id": "v1b2n3m4k5j6h7g8",
    "original_url": "https://mistymountains.io/signup/early-access"
  }
]`

var bodyCreateShortURLJSONListSUCCESSANDEXIST string = `[
  {
    "correlation_id": "a1b2c3d4e5f6g7h8",
    "original_url": "https://astrodata.io/reports/yearly-forecast"
  },
  {
    "correlation_id": "z9y8x7w6v5u4t3s2",
    "original_url": "https://bluemelon.net/gallery/9f3d1"
  },
  {
    "correlation_id": "m1n2b3v4c5x6z7l8",
    "original_url": "https://cloudyfox.org/blogs/hidden-pathways"
  },
  {
    "correlation_id": "q1w2e3r4t5y6u7i8",
    "original_url": "https://lunarbyte.com/downloads/setup_v2.1.exe"
  },
  {
    "correlation_id": "p0o9i8u7y6t5r4e3",
    "original_url": "https://pinetreehub.com/shop/category/garden-tools"
  },
  {
    "correlation_id": "n7m6b5v4c3x2z1a0",
    "original_url": "https://silentowl.net/articles/night-sky-mysteries"
  },
  {
    "correlation_id": "l8k7j6h5g4f3d2s1",
    "original_url": "https://crimsoncactus.com/store/new-arrivals"
  },
  {
    "correlation_id": "w3e2r1t0y9u8i7o6",
    "original_url": "https://silverpine.app/blogs/productivity-hacks"
  },
  {
    "correlation_id": "k2j3h4g5f6d7s8a9",
    "original_url": "https://greenglass.io/tools/ai-image-enhancer"
  },
  {
    "correlation_id": "b9n8m7v6c5x4z3l2",
    "original_url": "https://sunsetvoyagers.org/community/travel-stories"
  }
]`

var bodyCreateShortURLJSONListInvalidJSONArrayENDSTRING = `[
  {
    "correlation_id": "a1b2c3d4e5f6g7h8",
    "original_url": "https://astrodata.io/reports/yearly-forecast"
  },
  {
    "correlation_id": "z9y8x7w6v5u4t3s2",
    "original_url": "https://bluemelon.net/gallery/9f3d1"
  },
  {
    "correlation_id": "m1n2b3v4c5x6z7l8",
    "original_url": "https://cloudyfox.org/blogs/hidden-pathways"
  },
  {
    "correlation_id": "q1w2e3r4t5y6u7i8",
    "original_url": "https://lunarbyte.com/downloads/setup_v2.1.exe"
  },
  {
    "correlation_id": "p0o9i8u7y6t5r4e3",
    "original_url": "https://pinetreehub.com/shop/category/garden-tools"
  },`

package main

import "github.com/markbates/pkger"

import "github.com/markbates/pkger/pkging/mem"


var _ = pkger.Apply(mem.UnmarshalEmbed([]byte(`1f8b08000000000000ffec90310bc2301484ffcbcda1a1b86516c4a1e2a27bdbbcb681264f5ed2a9e4bf4bb33829dad9f9eee38e6f850b03479815a34bd3d2553d7bedbb992de99103272ae1d1090cf42d92c457aca3f4fa3da770f60f96746dd304f36140a161bbccb42d7d536e5d8049b290da7becc40ddb9f413d72e5d916fe4e121d0718d4557d4056b8b49e60e0b777392b0c6e2ef2b2c244427f8d7b343e010000ffff010000ffffcd4902f5a1020000`)))

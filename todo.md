#TODO

- [ ] use env variables instead of hardcoded config.

## (Premature) optimizations
- minimize string to []byte conversions
- use strings fro indexing and searching and []byte for concating.
- prefer appendFormat to Format. See https://segment.com/blog/allocation-efficiency-in-high-performance-go-services/; https://adamdrake.com/faster-command-line-tools-with-go.html
- 
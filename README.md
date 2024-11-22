# object cache action

This action is used to cache dependencies and build outputs with the cache object service to improve workflow execution time.

### Inputs

* `key` - An explicit key for a cache entry. See [creating a cache key](#creating-a-cache-key).
* `path` - A list of files, directories, and wildcard patterns to cache and restore. See [`@actions/glob`](https://github.com/actions/toolkit/tree/main/packages/glob) for supported patterns.
* `restore-keys` - An ordered multiline string listing the prefix-matched keys, that are used for restoring stale cache if no cache hit occurred for key.
* `endpoint` - The endpoint to use for the cache object service.

### Outputs

* `cache-hit` - A string value to indicate an exact match was found for the key.
  * If there's a cache hit, this will be 'true' or 'false' to indicate if there's an exact match for `key`.
  * If there's a cache miss, this will be an empty string.

# License

This application is released under Apache 2.0 license and is copyright [Mark Wolfe](https://www.wolfe.id.au).
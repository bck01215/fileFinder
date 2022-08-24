# fileFinder

## What it does?

This file finds all files on a filesystem and sorts them by size. It only returns the largest files that you specify

## How to use it

You can specify any directory and it will find every file in the directory and subdirectories if the subdirectories are on the same filesystem.

Example use:

```bash
./fileFinder --mount=/ -l 25
```

### ToDo

- [x] Remove iteration over items in channel bottleneck

### Author Information

Authors:

- [Brandon Kauffman](mailto:bck01215@gmail.com)

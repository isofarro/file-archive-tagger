# File Archive Tagging (FART)

## Architecture

FART is a simple CLI tool that allows you to tag files in a directory. It's designed to be used in a way that's similar to git. But it stores everything in a single SQLite database file called `.fart`.

FART then allows you to add tags to files. It then allows you to search for files by tags. It also allows you to check whether you already have a file by comparing the hash of the file's contents.

By default FART supports the taxonomy of `tags`. But you can create your own taxonomies, like `authors` or `projects` or `categories`.

## Development notes and guidelines

* Keep the database form in clean third normal form.
* Keep the code simple and easy to read and understand.
* Separate the code into modules with a logical structure, e.g. the database module, the cli module, the taxonomy module, etc.


## Using FART

It has a similar interface to git, but it's about associating metadata to files.

> fart init

Creates a file in the current directory called  `.fart` (if it doesn't already exist). This file is an SQLite database. This database stores all the data associated with this app.

It then recurses through the current directory adding the files it finds to the database (not the directories). The important data is the filename, the path (from the current directory), the hash of the file contents, the size and last modified date.

> fart add .
> fart add my-file.pdf
> fart add 2025/my-file.pdf
> fart add 2025/*.pdf

Adds files to the database, their metadata and a hash of the file's contents.

> fart tag 2025/my-file.pdf 2025-ideas

This tags the file with the tax `2025-ideas`

> fart taxonomy init author
> fart taxonomy init series

This creates a new taxonomies `author` and `series`.

> fart tag --author "Jordan, Robert" books/eye-of-the-world.pdf
> fart tag --series "The Wheel of Time" books/eye-of-the-world.pdf

These two commands add Robert Jordan as the author and that it's part of The Wheel of Time series of books.

> fart search --series "The Wheel of Time"

Returns all files that are part of the series The Wheel of Time.

> fart check ../incoming/random-file.pdf

Checks whether the file exists in the database, using the hash of the file's contents.

> fart stage ../incoming-files/

Sets `../incoming-files/` as the stage directory

> fart check

Checks all the files in the stage directory and reports whether they already exist in the database. Again, using the hash of the file's contents as the comparator.

> fart verify
> fart verify my-dir/
> fart verify my-other-dir/*.pdf

Verifies that all the files matched by the glob pattern exist in the database. The output is a list of files that are not in the database, or in the database and not in the directory, whether it's because they are new, or deleted, or because the file has been renamed or moved to a different location.

When run without parameters, it verifies all the files from the current directory and its sub-directories.


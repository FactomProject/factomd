# ProcessedBlockParsing

If `-wrproc` is turned on, the processed dbstates are written to disk in a rotating fashion. This tooll will allow you to open these files to look at.

To compare directories, you can see this little tutorial: https://imgur.com/a/MSwvNPh

```
# Directory being a collection of directories containing *.block files
ProcessedBlockParsing -d DIRECTORY

# A single file
ProcessedBlockParsing -f FILE.block
```
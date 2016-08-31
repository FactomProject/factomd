## Compiling the Static files into the binary.
 - go get github.com/FactomProject/staticfiles
 - Run controlPanel/compile.sh from within controlPanel directory
  - 'files/statics/statics.go' and 'files/templates/templates.go' are generated and should not be manually edited.

## Making Changes
 - Changes to javascript can me made in 'Web/uncompressedStatics'
  - When done, minify the js and replace the files located in 'Web/statics', then compile using 'compile.sh'

- After all changes, compile.sh must be run



 ### Notes:
 - If you add a third directory into the Web folder, custom management
 must be added in '/controlPanel/files/general.go' and 'compile.sh' must
 be adjusted.

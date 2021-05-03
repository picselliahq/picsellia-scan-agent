echo Release version ?
read version


go build -o=builds/picsell 
chmod 777 builds/picsell    
tar -zcvf tarbuilds/picsell_${version}_linux_x86_64.tar.gz builds/picsell 
rm builds/picsell 

env GOOS=windows GOARCH=amd64 go build -o=builds/picsell.exe
chmod 777 builds/picsell.exe    
zip tarbuilds/picsell_${version}_windows_amd64.zip builds/picsell.exe
rm builds/picsell.exe 

env GOOS=windows GOARCH=386 go build -o=builds/picsell.exe
chmod 777 builds/picsell.exe    
zip tarbuilds/picsell_${version}_windows_386.zip builds/picsell.exe
rm builds/picsell.exe 


env GOOS=darwin GOARCH=amd64 go build -o=builds/picsell
chmod 777 builds/picsell    
tar -zcvf tarbuilds/picsell_${version}_picsell_darwin_amd64.tar.gz builds/picsell 
rm builds/picsell 

env GOOS=darwin GOARCH=386 go build -o=builds/picsell
chmod 777 builds/picsell    
tar -zcvf tarbuilds/picsell_${version}_picsell_darwin_386.tar.gz builds/picsell 
rm builds/picsell 



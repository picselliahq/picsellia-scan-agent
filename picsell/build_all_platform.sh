echo Release version ?
read version


go build -o=builds/picsell 
chmod 777 builds/picsell    

cd builds 
ls
tar -zcvf ../tarbuilds/picsell_${version}_linux_x86_64.tar.gz picsell 
rm picsell 

cd ..
env GOOS=windows GOARCH=amd64 go build -o=builds/picsell.exe
chmod 777 builds/picsell.exe 

cd builds
zip ../tarbuilds/picsell_${version}_windows_amd64.zip picsell.exe 
rm picsell.exe 

cd ..
env GOOS=windows GOARCH=386 go build -o=builds/picsell.exe
chmod 777 builds/picsell.exe    

cd builds
zip ../tarbuilds/picsell_${version}_windows_386.zip picsell.exe
rm picsell.exe 

cd ..

env GOOS=darwin GOARCH=amd64 go build -o=builds/picsell
chmod 777 builds/picsell    

cd builds
tar -zcvf ../tarbuilds/picsell_${version}_picsell_darwin_amd64.tar.gz picsell 
rm picsell 

cd ..

env GOOS=darwin GOARCH=386 go build -o=builds/picsell
chmod 777 builds/picsell    

cd builds 
tar -zcvf ../tarbuilds/picsell_${version}_picsell_darwin_386.tar.gz picsell 
rm picsell

cd ..



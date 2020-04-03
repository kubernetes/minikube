for file in $(find . -type f -name "*.go")
do
    if [[ -f $file ]]; then
        gofmt -w $file      
        goimports -w $file      
    fi
done
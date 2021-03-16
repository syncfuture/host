#git tag | foreach-object -process { git push origin --delete $_ }
#git tag | foreach-object -process { git tag -d $_ }
git push origin --delete v1.3.14
git tag -d v1.3.14
git tag v1.3.15
git push
git push --tags
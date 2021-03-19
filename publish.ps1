$OldTag = "v1.3.41"
$NewTag = "v1.3.42"
#git tag | foreach-object -process { git push origin --delete $_ }
#git tag | foreach-object -process { git tag -d $_ }
git push origin --delete $OldTag
git tag -d $OldTag
git tag $NewTag
git push
git push --tags
#git tag | foreach-object -process { git push origin --delete $_ }
#git tag | foreach-object -process { git tag -d $_ }

# $OldTag = "v1.3.62"
$NewTag = "v1.6.0"
# git push origin --delete $OldTag
# git tag -d $OldTag
git tag $NewTag
git push
git push --tags
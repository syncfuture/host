git tag | foreach-object -process { git push origin --delete $_ }
git tag | foreach-object -process { git tag -d $_ }
<<<<<<< HEAD
git tag v1.1.4
=======
git tag v1.1.7
>>>>>>> 0c6e26d5f9440a7a199dbf5a07c9a77e3060fe36
git push
git push --tags
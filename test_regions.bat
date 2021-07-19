echo on
for /l %%x in (%1, -1, 0) do (
   pulumi config set region0Servers %x
     for /l %%y in (%1, -1, 0) do (
         pulumi config set region1Servers %y
         for /l %%z in (%1, -1, 0) do (
             pulumi config set region2Servers %z
             IF NOT %x == 0 (IF NOT %y == 0 (IF NOT %z == 0 (echo %x %y %z)))
             )
         )
     )
 )
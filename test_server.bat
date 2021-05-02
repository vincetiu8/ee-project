@echo off
pulumi stack select zone-%1
Set /a test_zone = (%1+%2) %% 13
pulumi config set testerZone %test_zone%
pulumi up --yes
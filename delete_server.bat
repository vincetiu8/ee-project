@echo off
pulumi stack select zone-%1
pulumi destroy --yes
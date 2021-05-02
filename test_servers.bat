@echo off
for /l %%x in (0, 1, 6) do test_server.bat %%x %1
time /T
timeout /t /nobreak 2000
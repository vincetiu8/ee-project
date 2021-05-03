for /l %%x in (%1, 1, 2) do (test_servers.bat %%x & timeout /t 1000)
for /l %%x in (0, 1, 2) do delete_server.bat %%x
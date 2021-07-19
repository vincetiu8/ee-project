import subprocess
import os
import time

regions = ["us-east-1", "ap-east-1", "eu-central-1"]

depth = int(input("How many servers do you want to set up per region? "))
for i in range(1, 3):
    os.system(f"pulumi config set hostRegion {i}")
    os.system(f"pulumi config set region{i}Servers 1")
    region1 = 0
    region2 = 2
    if i == 2:
        region2 = 1
    for j in range(depth, -1, -1):
        os.system(f"pulumi config set region{region1}Servers {j}")
        for k in range(depth, -1, -1):
            os.system(f"pulumi config set region{region2}Servers {k}")
            for l in range(3):
                name = f"test-{j}-"
                if i == 1:
                    name += f"1-{k}"
                else:
                    name += f"{k}-1"
                try:
                    result = subprocess.check_call(f"aws s3 ls s3://ipfs-output-bucket/{name}-from-{regions[i]}-to-{regions[l]}")
                    continue
                except subprocess.CalledProcessError:
                    os.system(f"pulumi config set testerRegion {l}")
                    os.system("pulumi up -y")
                while True:
                    try:
                        result = subprocess.check_call(f"aws s3 ls s3://ipfs-output-bucket/{name}-from-{regions[i]}-to-{regions[l]}")
                        break
                    except subprocess.CalledProcessError:
                        print("couldn't find file...")
                        time.sleep(120)
os.system("pulumi destroy -y")
# Install bootdev cli
go install github.com/bootdotdev/bootdev@latest

# sqlite
sudo apt update
sudo apt install sqlite3

Email: admin@tubely.com
Password: password

# localstack to simulate aws environment 
docker compose up -d

# Install aws cli 
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install -i /usr/local/aws-cli -b /usr/local/bin
sudo -i
aws --version

nano ~/.aws/config

[default]
region = us-east-1
output = json
endpoint_url = http://localhost:4566

aws sts get-caller-identity
aws iam create-group --group-name managers
aws iam attach-group-policy \
    --group-name managers \
    --policy-arn arn:aws:iam::aws:policy/AdministratorAccess
aws iam create-user --user-name hktikhin
aws iam create-access-key --user-name hktikhin
aws iam add-user-to-group \
    --user-name hktikhin \
    --group-name managers
aws s3api create-bucket \
    --bucket tubely-82931 \
    --region us-east-1
aws s3api delete-public-access-block \
    --bucket tubely-82931
aws s3api put-bucket-policy \
    --bucket tubely-82931 \
    --policy file://policy.json
aws s3 cp boots-image-horizontal.png s3://tubely-82931/
https://sturdy-palm-tree-9wx4q976p99cwj-4566.app.github.dev/tubely-82931/boots-image-horizontal.png

aws s3 ls s3://tubely-82931 > /tmp/bucket_contents.txt

go get github.com/aws/aws-sdk-go-v2/service/s3 github.com/aws/aws-sdk-go-v2/config

curl -v -X POST http://127.0.0.1:8091/api/video_upload/1ae2eead-41c3-4249-a0fd-dcd9bd96f37f \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ0dWJlbHktYWNjZXNzIiwic3ViIjoiMDUxNGI1YzMtMTJkNC00N2QwLTgyMTgtODYyNjlmODFhMzE3IiwiZXhwIjoxNzc4MzEyMTAyLCJpYXQiOjE3NzU3MjAxMDJ9.6pK6MRaxAjKnZw2twgT9wO6N-j0hcszZAG1MTj46Kr4" \
  -F "video=@samples/boots-video-horizontal.mp4;type=video/mp4"

  https://sturdy-palm-tree-9wx4q976p99cwj-8091.app.github.dev/
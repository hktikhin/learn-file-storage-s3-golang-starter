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
aws iam create-policy \
    --policy-name manager-from-home \
    --policy-document file://manager-policy.json
aws iam list-attached-group-policies --group-name managers
aws iam get-policy-version \
    --policy-arn arn:aws:iam::000000000000:policy/manager-from-home \
    --version-id v1
aws iam attach-group-policy \
    --group-name managers \
    --policy-arn arn:aws:iam::000000000000:policy/manager-from-home
aws iam create-user --user-name hktikhin
aws iam create-access-key --user-name hktikhin
aws iam list-access-keys --user-name hktikhin
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

curl -v -X POST http://127.0.0.1:8091/api/video_upload/2ae356b7-4384-44df-aa56-4de9dfaac692 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ0dWJlbHktYWNjZXNzIiwic3ViIjoiYWQyMzNkZDgtMjYwZS00MTE1LWI4NTYtYTgwZTU3M2RiODBlIiwiZXhwIjoxNzc4ODMxMjUxLCJpYXQiOjE3NzYyMzkyNTF9.QMHWTrB587FJzcDywRYWxhhKuNiE5xtMuOLLd5rHPh0" \
  -F "video=@samples/boots-video-horizontal.mp4;type=video/mp4"

curl -v -X POST http://127.0.0.1:8091/api/video_upload/bd8aa302-ad05-48f2-8a3d-072cb304648a \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ0dWJlbHktYWNjZXNzIiwic3ViIjoiMDUxNGI1YzMtMTJkNC00N2QwLTgyMTgtODYyNjlmODFhMzE3IiwiZXhwIjoxNzc4MzEyMTAyLCJpYXQiOjE3NzU3MjAxMDJ9.6pK6MRaxAjKnZw2twgT9wO6N-j0hcszZAG1MTj46Kr4" \
  -F "video=@samples/boots-video-vertical.mp4;type=video/mp4"

aws s3api head-object --bucket tubely-82931 --key RZBi5Srqjh-ka-0SOHDXgXb5QKgZupUFhgYD7wks_K8.mp4 > /tmp/object_metadata.txt

aws s3api put-object --bucket tubely-82931 --key backups/
aws s3 sync ./samples s3://tubely-82931/backups/
aws s3 ls s3://tubely-82931/backups/
aws s3 ls s3://tubely-82931/backups/ > /tmp/s3_listing.txt

sudo apt install ffmpeg
ffprobe -version
ffmpeg -version

ffprobe -v error -print_format json -show_streams samples/boots-video-horizontal.mp4

# Improve Security flow
aws s3api create-bucket \
    --bucket tubely-private-99281 \
    --region us-east-1

aws s3api put-public-access-block \
    --bucket tubely-private-99281 \
    --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

aws s3 ls s3://tubely-private-99281

aws s3api get-public-access-block --bucket tubely-private-99281
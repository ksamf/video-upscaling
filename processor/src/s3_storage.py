from contextlib import asynccontextmanager
import os
from aiobotocore.session import get_session

from config import settings
from botocore.exceptions import ClientError


class S3Client:
    def __init__(
        self,
        access_key: str,
        secret_key: str,
        endpoint_url: str,
        bucket_name: str,
    ):
        self.config = {
            "aws_access_key_id": access_key,
            "aws_secret_access_key": secret_key,
            "endpoint_url": "https://" + endpoint_url,
        }
        self.bucket_name = bucket_name
        self.session = get_session()

    @asynccontextmanager
    async def get_client(self):
        async with self.session.create_client("s3", **self.config) as client:
            yield client

    async def upload_file(
        self,
        folder_name: str,
        file_path: str,
    ):
        """
        Uploads a file to the specified S3 bucket and folder.

        Args:
            folder_name (str): Folder in the S3 bucket where the file will be uploaded.
            file_path (str): Path to the local file to be uploaded.
        """
        try:
            async with self.get_client() as client:
                with open(file_path, "rb") as f:
                    await client.put_object(
                        Bucket=self.bucket_name,
                        Key=f"{folder_name}/{os.path.basename(file_path)}",
                        Body=f,
                    )
            print(f"File '{file_path}' uploaded successfully.")
        except ClientError as e:
            print(f"Error uploading file: {e}")

    async def get_file(self, object_name, destination_path):
        """
        Downloads a file from the specified S3 bucket.
        Args:
            object_name (str): The name of the object to download from S3.
            destination_path (str): The local path where the file will be saved.
        """
        async with self.get_client() as client:
            response = await client.get_object(Bucket=self.bucket_name, Key=object_name)
            data = await response["Body"].read()
            with open(destination_path, "wb") as file:
                file.write(data)
            print(f"File {object_name} downloaded to {destination_path}")

    async def file_exists(self, key: str) -> bool:
        """
        Checks if a file exists in the specified S3 bucket.
        Args:
            key (str): The key of the object to check in S3.
        Returns:
            bool: True if the file exists, False otherwise.
        """
        async with self.get_client() as client:
            try:
                await client.head_object(Bucket=self.bucket_name, Key=key)
                return True
            except ClientError as e:
                if e.response["Error"]["Code"] == "403":
                    print("Access denied.")
                elif e.response["Error"]["Code"] == "404":
                    print("File does not exist.")
                else:
                    raise


s3_client = S3Client(
    access_key=settings.S3_ACCESS_KEY_ID,
    secret_key=settings.S3_SECRET_ACCESS_KEY,
    endpoint_url=settings.S3_ENDPOINT_URL,
    bucket_name=settings.S3_BUCKET_NAME,
)

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    APP_HOST: str
    APP_PORT: int
    APP_DEBUG: bool

    S3_ACCESS_KEY_ID: str
    S3_SECRET_ACCESS_KEY: str
    S3_ENDPOINT_URL: str
    S3_BUCKET_NAME: str

    UPSCALE_3D_PATH: str
    UPSCALE_2D_PATH: str
    TRANSCRIBE_PATH: str
    TRANSLATE_PATH: str

    model_config = SettingsConfigDict(env_file=".env")


settings = Settings()

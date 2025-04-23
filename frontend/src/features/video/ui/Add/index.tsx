import { FC, useState } from "react";
import { Button } from "@heroui/react";
import { Modal, ModalBody, ModalContent, ModalFooter, ModalHeader, useDisclosure } from "@heroui/modal"
import { useTranslation } from "react-i18next";
import { UploadInput } from "@shared/ui";
import { VideoPlayer } from "../Player";
import MinusCircleIcon from "@shared/assets/icons/minus-cirlce.svg?react"
import { useUploadVideo } from "@features/video/hooks/useUploadVideo";

const emptyFormState = {
    filepath: ""
}

export const AddVideoModal: FC = () => {
    const { t } = useTranslation()
    const { isOpen, onOpen, onOpenChange, onClose } = useDisclosure();
    const [formState, setFormState] = useState<{ filepath: string }>(emptyFormState);
    const { upload, isLoading, isError, error } = useUploadVideo();

    const handleUpload = async () => {
        if (formState.filepath.length === 0) return;
        const res = await upload(formState.filepath);
        if (typeof res !== "string") {
            // do somthing, for example add to store
            onClose()
        }
    }

    return (
        <>
            <Button onPress={onOpen} color="primary">
                {t("header.load.trigger")}
            </Button>
            <Modal isOpen={isOpen} onOpenChange={onOpenChange} size="xl">
                <ModalContent>
                    <ModalHeader>{t("header.load.trigger")}</ModalHeader>
                    <ModalBody>
                        {!formState.filepath ? (
                            <UploadInput
                                onChange={data => setFormState(prev => ({
                                    ...prev,
                                    filepath: data.filepath
                                }))}
                                accept="video/*"
                                multiple={false}
                            />
                        ) : (
                            <div className={`relative flex flex-col gap-2 items-end`}>
                                <VideoPlayer src={formState.filepath} />
                                <Button
                                    size="sm"
                                    color="danger"
                                    className={`font-semibold`}
                                    onPress={() => setFormState(emptyFormState)}
                                    endContent={<MinusCircleIcon className={`stroke-white w-5 h-5`} />}
                                >
                                    {t("header.load.delete_file")}
                                </Button>
                            </div>
                        )}
                        {isError && (
                            <p className={`text-red-600`}>
                                {error}
                            </p>
                        )}
                    </ModalBody>
                    <ModalFooter>
                        <Button
                            onPress={handleUpload}
                            color="primary"
                            variant="shadow"
                            fullWidth
                            size="lg"
                            isDisabled={!formState.filepath}
                            className={`font-semibold`}
                            isLoading={isLoading}
                        >
                            {t("header.load.load")}
                        </Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
        </>
    )
}
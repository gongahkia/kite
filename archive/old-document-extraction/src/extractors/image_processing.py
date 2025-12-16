from PIL import Image, ImageFilter, ImageEnhance

class ImageProcessing:
    @staticmethod
    def deskew(image):
        gray = image.convert('L')
        return gray

    @staticmethod
    def denoise(image):
        return image.filter(ImageFilter.MedianFilter(size=3))

    @staticmethod
    def adjust_contrast(image, factor=1.5):
        enhancer = ImageEnhance.Contrast(image)
        return enhancer.enhance(factor)

    @staticmethod
    def preprocess_image(image):
        image = ImageProcessing.deskew(image)
        image = ImageProcessing.denoise(image)
        image = ImageProcessing.adjust_contrast(image)
        return image
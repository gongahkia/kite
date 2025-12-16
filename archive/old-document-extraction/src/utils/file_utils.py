import os

def list_files(directory, extensions=None):
    files = []
    for root, _, filenames in os.walk(directory):
        for filename in filenames:
            if not extensions or filename.lower().endswith(tuple(extensions)):
                files.append(os.path.join(root, filename))
    return files

def is_supported_file(filename):
    supported = ('.pdf', '.docx')
    return filename.lower().endswith(supported)
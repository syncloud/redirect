import os
import shutil
import argparse

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='This is redirect configuration tool')
    parser.add_argument('--configuration', dest='configuration')
    args = parser.parse_args()

    configuration = args.configuration

    look_for = '_'+configuration+'.'

    for dirpath, dnames, fnames in os.walk("."):
        for f in fnames:
            if f.startswith(look_for):
                src = os.path.join(dirpath, f)
                dst = os.path.join(dirpath, f.replace(look_for, ''))
                print('Copying configuration file {} to {}'.format(src, dst))
                shutil.copyfile(src, dst)
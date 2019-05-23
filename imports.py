import argparse
import xdrlib
import ast
from ast import parse, walk

def main():
    parser = argparse.ArgumentParser(description='Find deprecated imports.')
    parser.add_argument('source', type=argparse.FileType('r'))
    args = parser.parse_args()
    root = parse(args.source.read())

    dead = set([
        "aifc",
        "asynchat",
        "asyncore",
        "audioop"
        "binhex",
        "cgi"
        "cgitb",
        "chunk",
        "crypt",
        "formatter",
        "fpectl",
        "imghdr",
        "imp",
        "macpath",
        "msilib",
        "nis",
        "nntplib",
        "ossaudiodev",
        "parser",
        "pipes",
        "smtpd",
        "sndhdr",
        "spwd",
        "sunau",
        "uu",
        "xdrlib",
    ])

    found = set([])

    for node in walk(root):
        if isinstance(node, ast.Import):
            if node.names[0].name in dead:
                found.add(node.names[0].name)
        if isinstance(node, ast.ImportFrom):
            if node.module in dead:
                found.add(node.module)

    if found:
        print(','.join(found))

if __name__ == "__main__":
    main()

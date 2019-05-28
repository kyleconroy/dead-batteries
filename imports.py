import argparse
import ast

from ast import parse, walk
from flask import Flask, jsonify, request

app = Flask(__name__)

DEAD = set([
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

@app.route("/", methods=["POST"])
def scan():
    try:
        with open(request.json['path']) as r:
            root = parse(r.read())
    except SyntaxError:
        return jsonify(imports=[], error="syntax-error")
    except UnicodeDecodeError:
        return jsonify(imports=[], error="unicode-decode-error")
    except ValueError:
        return jsonify(imports=[], error="value-error")

    found = set([])
    for node in walk(root):
        if isinstance(node, ast.Import):
            if node.names[0].name in DEAD:
                found.add(node.names[0].name)
        if isinstance(node, ast.ImportFrom):
            if node.module in DEAD:
                found.add(node.module)

    return jsonify(imports=list(found))

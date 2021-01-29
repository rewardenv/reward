import sys, os
import sphinx_rtd_theme

import recommonmark
from recommonmark.transform import AutoStructify


f = open("../VERSION.txt", "r")
release = f.read()

project = 'Reward'
copyright = '2021, Janos Miko - mixe3y <janos.miko@itg.cloud>'
author = 'Janos Miko - mixe3y <janos.miko@itg.cloud>'

extensions = [
  'recommonmark',
  'sphinx_rtd_theme',
  'sphinx_copybutton',
  'sphinx_markdown_tables',
]

templates_path = ['_templates']
exclude_patterns = ['_build', 'Thumbs.db', '.DS_Store']
html_theme = "sphinx_rtd_theme"
html_static_path = ['_static']

def setup(app):
    app.add_config_value('recommonmark_config', {
        'auto_toc_tree_section': ['Table of Contents'],
        'enable_math': False,
        'enable_inline_math': False,
        'enable_eval_rst': True,
    }, True)
    app.add_transform(AutoStructify)


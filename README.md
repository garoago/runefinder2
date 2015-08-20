# runefinder2
Utility to find Unicode characters by words in the name

Sample uses:

```
$ runefinder2 cat face
U+1F431 🐱 	CAT FACE
U+1F638 😸 	GRINNING CAT FACE WITH SMILING EYES
U+1F639 😹 	CAT FACE WITH TEARS OF JOY
U+1F63A 😺 	SMILING CAT FACE WITH OPEN MOUTH
U+1F63B 😻 	SMILING CAT FACE WITH HEART-SHAPED EYES
U+1F63C 😼 	CAT FACE WITH WRY SMILE
U+1F63D 😽 	KISSING CAT FACE WITH CLOSED EYES
U+1F63E 😾 	POUTING CAT FACE
U+1F63F 😿 	CRYING CAT FACE
U+1F640 🙀 	WEARY CAT FACE
10 characters found

$ runefinder2 cat face smiling
U+1F638 😸 	GRINNING CAT FACE WITH SMILING EYES
U+1F63A 😺 	SMILING CAT FACE WITH OPEN MOUTH
U+1F63B 😻 	SMILING CAT FACE WITH HEART-SHAPED EYES
3 characters found
```

## Credits

This is a port of the Python [charfinder](https://github.com/fluentpython/example-code/tree/master/18-asyncio/charfinder) utilities created for the [Fluent Python](http://shop.oreilly.com/product/0636920032519.do) book. 

Go development was done during [Garoa Gophers](https://garoa.net.br/wiki/Garoa_Gophers) meetings with:

* Alexandre Souza ([@alexandre](https://github.com/alexandre/))
* Andrews Medina ([@andrewsmedina](https://github.com/andrewsmedina/))
* João "JC" Martins ([@jcmartins](https://github.com/jcmartins))
* Luciano Ramalho ([@ramalho](https://github.com/ramalho/))
* Michael Howard

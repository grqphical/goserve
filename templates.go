package main

const DirectoryTemplate string = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>$DIRECTORY$</title>
    <style>
        th, td {
            padding-top: 10px;
            padding-bottom: 10px;
            padding-left: 10px;
            padding-right: 50px;
        }
    </style>
</head>
<body>
    <h1>Contents of $DIRECTORY$</h1>
    <table>
        <tr>
            <th>File Name</th>
            <th>Size</th>
        </tr>
        $LINKS$
    </table>
    
</body>
</html>`

const NotFoundTemplate string = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File Not Found</title>
</head>
<body>
    <h1>404 Not Found</h1>
    <p>Server could not find the file you were looking for</p>
</body>
</html>`

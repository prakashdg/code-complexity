### Complexity measurement tools provide several pieces of information. They help to:

-   locate suspicious areas in unfamiliar code
-   get an idea of how much effort may be required to understand that code
-   get an idea of the effort required to test a code base
-   provide a reminder to yourself. You may see what you’ve written as obvious, but others may not. It is useful to have a hint about what code may seem harder to understand by others, and then decide if some rework may be in order.

### Steps to generate complexity report

git checkout source-branch

complexity --histogram --score --thresh=3 *.c > /src/source_file_complexity.txt

git checkout target-branch

complexity --histogram --score --thresh=3 *.c > /src/target_file_complexity.txt


complexity_generator -sourceComplexity=/src/source_file_complexity.txt -targetComplexity=/src/target_file_complexity.txt -mrID=$gitMR_ID -changesetFile=/src/code_changes_diff


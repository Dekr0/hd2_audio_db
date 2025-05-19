- Be cautions when using `map` due to its potential memory leak.
- Document your function call and method call as you write them. This is 
especially important when a function / method call that return some values along 
with error. Specify what situation it's safe to use those values and.
- Prefer correctness over speed when using pointers. However, `struct` and any 
other non-primitive data type is a different story.
- To cope with potential `nil` pointer in a very heavy data oriented logic, 
enforce "never trust" policies. Double check incoming pointer
- Avoid continuing program logic when error occur. This can potentially introduce 
more bug in the future because you allow some degree of invariants not to be 
held.

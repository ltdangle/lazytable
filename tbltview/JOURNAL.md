# Journal
## 19 June, 2024
This is a second attempt to integrate lua as the formula/scripting language. In the first attempt I was able to integrate lua (was bitten by the registry size issue, wtf?). Created a lua script with global ```lua Data = {}``` and global functions (getters, setters, mutators) on this global variable. For every request to the datastore a corresponding lua function was called, which worked, but I can already tell that it is slowing down the interaction. On the second thought, the passing every data request to lua is an overkill: I can mirror go data structure to lua only on data mutation and formula calculation.
For this to work need to:
1. Have formulas be calculated per data mutation, not per cell request (Cell.Calculate()). GetCell() method should just return the formula result if it is set.
2. Identify current data state mutators and have it mirror state to lua data store. Mirror logic should execute any lua formulas and store result of the calculation in go data structure.

## 10 July, 2024
The project is scrapped if favor of using sc-im. sc-im already has vim bindings, lua scripting, undo/redo, filter, sort, and is basically what I was trying to build. Higher ROI would be achieved by focusing on mastering sc-im. 
Sc-im is not without its flaws, but it scratches an itch if you don't want to use gui spreadsheets.
Time wasted or invested? Throughly enjoyed building/playing with it and some lessons learned, even if they are hard to articulate. 



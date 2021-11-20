# Design patterns

## Chain of Responsibility

Consisting of a source of command objects and a series of processing objects.

![diagram](pic/uml_chain_of_responsibility.jpg)

## Memento

Provides the ability to restore an object to its previous state (undo via rollback). It's operates on a single object.

* **Orginator:** The originator is some object that has an internal state.
* **Memento:** an opaque object (is a data type whose concrete data structure is not defined in an interface) which the caretaker cannot, or should not, change.
* **CareTacker:** Asks the originator first, for a memento object 

The memento object is hidden inside the Originator and can't be accessed from outsite. Only the originator that created a memento is allowed to access it.

Methods:

* Originator
	* createMemento(): Return new mutable object
	* restore(memento)
* Memento
	* getState()
	* setState()
	
![diagram](pic/uml_memento.jpeg)

## Observer

The observer pattern is a software design pattern in which an object, named the **subject**, maintains a list of its dependents, **called observers, and notifies them automatically of any state changes**.

![diagram](pic/uml_observer.jpg)

### Reference
* Wikipedia
	* [COR](https://en.wikipedia.org/wiki/Chain-of-responsibility_pattern)
	* [Memento](https://en.wikipedia.org/wiki/Memento_pattern)
	* [Observer](https://en.wikipedia.org/wiki/Observer_pattern)
* [Refactoring GURU](https://refactoring.guru/design-patterns/catalog)
* [Design Patterns. Elements of Reusable Object-Oriented Software](https://www.amazon.de/-/en/Erich-Gamma/dp/0201633612)
class Person {
  fullName: string;
  constructor (public firstName: string, public lastName: string) {
    this.fullName = `${firstName} ${lastName}`
  }

  greeting (): string {
    return `Hello~I'm ${this.fullName}~Nice to meet you~`
  }
}

interface Teenager extends Person {
  age: number;
}

class Student extends Person implements Teenager {
  greeting (): string {
    return `Hello~My name is ${this.fullName}, I'm ${this.age} years old`
  }

  constructor (public firstName: string, public lastName: string, public age: number) {
    super(firstName, lastName)
  }
}

const father = new Person('Li', 'Lei')
const user = new Student('Li', 'Mei', 8)

document.body.innerHTML = `
father: ${father.greeting()}<br/>
me: ${user.greeting()}
`

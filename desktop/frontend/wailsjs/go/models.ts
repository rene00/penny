export namespace model {
	
	export class Account {
	    id: number;
	    name: string;
	    parent_id: number;
	    account_type_id: number;
	    entity_id: number;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.parent_id = source["parent_id"];
	        this.account_type_id = source["account_type_id"];
	        this.entity_id = source["entity_id"];
	    }
	}
	export class AccountType {
	    id: number;
	    name: string;
	    parent_id: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.parent_id = source["parent_id"];
	    }
	}
	export class Entity {
	    id: number;
	    name: string;
	    entity_type_id: number;
	
	    static createFrom(source: any = {}) {
	        return new Entity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.entity_type_id = source["entity_type_id"];
	    }
	}
	export class Split {
	    id: number;
	    transactions_id: number;
	    account_id: number;
	    value_num: number;
	    value_denom: number;
	
	    static createFrom(source: any = {}) {
	        return new Split(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.transactions_id = source["transactions_id"];
	        this.account_id = source["account_id"];
	        this.value_num = source["value_num"];
	        this.value_denom = source["value_denom"];
	    }
	}
	export class AccountEntity {
	    id: number;
	    account_id: number;
	    entity_id: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountEntity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.account_id = source["account_id"];
	        this.entity_id = source["entity_id"];
	    }
	}
	export class AccountAttribute {
	    id: number;
	    account_id: number;
	    name: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountAttribute(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.account_id = source["account_id"];
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	export class accountR {
	    AccountType?: AccountType;
	    Entity?: Entity;
	    AccountAttributes: AccountAttribute[];
	    AccountEntities: AccountEntity[];
	    Splits: Split[];
	
	    static createFrom(source: any = {}) {
	        return new accountR(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AccountType = this.convertValues(source["AccountType"], AccountType);
	        this.Entity = this.convertValues(source["Entity"], Entity);
	        this.AccountAttributes = this.convertValues(source["AccountAttributes"], AccountAttribute);
	        this.AccountEntities = this.convertValues(source["AccountEntities"], AccountEntity);
	        this.Splits = this.convertValues(source["Splits"], Split);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AccountTypeAttribute {
	    id: number;
	    account_type_id: number;
	    name: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountTypeAttribute(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.account_type_id = source["account_type_id"];
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	export class accountTypeR {
	    Accounts: Account[];
	    AccountTypeAttributes: AccountTypeAttribute[];
	
	    static createFrom(source: any = {}) {
	        return new accountTypeR(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Accounts = this.convertValues(source["Accounts"], Account);
	        this.AccountTypeAttributes = this.convertValues(source["AccountTypeAttributes"], AccountTypeAttribute);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Transaction {
	    id: number;
	    currency_id: number;
	    entity_id: number;
	    memo: string;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new Transaction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.currency_id = source["currency_id"];
	        this.entity_id = source["entity_id"];
	        this.memo = source["memo"];
	        this.date = source["date"];
	    }
	}
	export class EntityAttribute {
	    id: number;
	    entity_id: number;
	    name: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new EntityAttribute(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.entity_id = source["entity_id"];
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	export class entityR {
	    Accounts: Account[];
	    AccountEntities: AccountEntity[];
	    EntityAttributes: EntityAttribute[];
	    Transactions: Transaction[];
	
	    static createFrom(source: any = {}) {
	        return new entityR(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Accounts = this.convertValues(source["Accounts"], Account);
	        this.AccountEntities = this.convertValues(source["AccountEntities"], AccountEntity);
	        this.EntityAttributes = this.convertValues(source["EntityAttributes"], EntityAttribute);
	        this.Transactions = this.convertValues(source["Transactions"], Transaction);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace types {
	
	export class AccountType {
	    id: number;
	    name: string;
	    id: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class Account {
	    id: number;
	    name: string;
	    description: string;
	    account_type: AccountType;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.account_type = this.convertValues(source["account_type"], AccountType);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class JSResp {
	    success: boolean;
	    msg: string;
	    data?: any;
	
	    static createFrom(source: any = {}) {
	        return new JSResp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.msg = source["msg"];
	        this.data = source["data"];
	    }
	}
	export class Split {
	    id: number;
	    value_num: number;
	    value_denom: number;
	    amount: string;
	    account: Account;
	
	    static createFrom(source: any = {}) {
	        return new Split(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.value_num = source["value_num"];
	        this.value_denom = source["value_denom"];
	        this.amount = source["amount"];
	        this.account = this.convertValues(source["account"], Account);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SplitResponse {
	    success: boolean;
	    msg: string;
	    data: Split;
	
	    static createFrom(source: any = {}) {
	        return new SplitResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.msg = source["msg"];
	        this.data = this.convertValues(source["data"], Split);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Transaction {
	    id: number;
	    memo: string;
	    date: string;
	    id: number;
	    name: string;
	    splits: Split[];
	    amount: string;
	
	    static createFrom(source: any = {}) {
	        return new Transaction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.memo = source["memo"];
	        this.date = source["date"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.splits = this.convertValues(source["splits"], Split);
	        this.amount = source["amount"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TransactionResponse {
	    msg: boolean;
	    success: string;
	    data: Transaction;
	
	    static createFrom(source: any = {}) {
	        return new TransactionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.msg = source["msg"];
	        this.success = source["success"];
	        this.data = this.convertValues(source["data"], Transaction);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}


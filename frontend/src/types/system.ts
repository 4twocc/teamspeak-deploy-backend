export interface System {
	id: number
	cpu: number
	memory: Partial<Memory>
	disk?: Partial<Disk>
	network?: Partial<Network>
	load?: Partial<Load>
	uptime?: number
	alert?: string | null
	timestamp: string
}

type Memory = {
	total: number
	used: number
	available: number
	usage: number
}

type Disk = {
	total: number
	used: number
	free: number
	usage: number
}

type Network = {
	rx_bytes: number
	tx_bytes: number
	rx_rate: number
	tx_rate: number
}

type Load = {
	load1: number
	load5: number
	load15: number
}

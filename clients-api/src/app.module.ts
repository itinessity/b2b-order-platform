import {Module} from '@nestjs/common'; import {ClientsController} from './clients/clients.controller'; import {ClientsService,MemoryClientsRepository} from './clients/clients.service'; import {HealthController} from './health/health.controller';
@Module({controllers:[ClientsController,HealthController],providers:[ClientsService,MemoryClientsRepository]}) export class AppModule{}

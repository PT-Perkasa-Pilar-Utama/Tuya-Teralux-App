package com.example.teraluxapp.di

import com.example.teraluxapp.data.repository.AuthRepository
import com.example.teraluxapp.data.repository.AuthRepositoryImpl
import com.example.teraluxapp.data.repository.TeraluxRepository
import com.example.teraluxapp.data.repository.TeraluxRepositoryImpl
import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
abstract class RepositoryModule {
    
    @Binds
    @Singleton
    abstract fun bindTeraluxRepository(
        teraluxRepositoryImpl: TeraluxRepositoryImpl
    ): TeraluxRepository
    
    @Binds
    @Singleton
    abstract fun bindAuthRepository(
        authRepositoryImpl: AuthRepositoryImpl
    ): AuthRepository
}
